
import * as vscode from 'vscode';

import * as path from 'path';
import * as fs from 'fs';
import * as shelljs from 'shelljs';
import * as download from 'download';
import * as crypto from 'crypto';

import * as config from './config';

// import squashVersionData = require('./squash.json');
import squashVersionData from './squash.json';


class BinariesSha {
    linux!: string;
    darwin!: string;
    win32!: string;
}
export class SquashInfo {
    version!: string;
    baseName!: string;
    binaries!: BinariesSha;
}



export function getSquashInfo(): SquashInfo {
    return <SquashInfo>squashVersionData;
}

interface SquashctlBinary {
    link: string;
    checksum: string;
}

function createSquashctlBinary(os: string, checksum: string): SquashctlBinary {
    let link = "https://github.com/solo-io/squash/releases/download/v" + getSquashInfo().version + "/" + getSquashInfo().baseName + "-" + os;
    console.log("downloading from: " + link);
    if ( config.get_conf_or("verbose", false) ) {
        vscode.window.showInformationMessage("downloading from: " + link);
    }
    return {
        link,
        checksum: checksum
    };
}

export function getSquashctl(): SquashctlBinary {
    // download the squash version for this extension
    var osver = process.platform;
    switch (osver) {
        case 'linux':
            return createSquashctlBinary("linux", getSquashInfo().binaries.linux);
        case 'darwin':
            return createSquashctlBinary("darwin", getSquashInfo().binaries.darwin);
        case 'win32':
            return createSquashctlBinary("windows.exe", getSquashInfo().binaries.win32);
        default:
            throw new Error(osver + " is current unsupported");
    }
}


export async function getremote(extPath: string): Promise<string> {
    let pathforbin = path.join(extPath, "binaries", getSquashInfo().version);
    let execpath = path.join(pathforbin, "squashctl");
    let ks = getSquashctl();

    // exit this early until release is smoothed out
    if (fs.existsSync(execpath)) {
        let exechash = await hash(execpath);
        // make sure its the one we expect:
        // this can happen on version updates.
        if (!hashesMatch(ks.checksum, exechash)) {
            // remove the bad binary.
            fs.unlinkSync(execpath);
        }
    }

    if (!fs.existsSync(execpath)) {
        let s = await vscode.window.showInformationMessage("Download Squash?", "yes", "no");
        if (s === "yes") {
            vscode.window.showInformationMessage("download started");
            shelljs.mkdir('-p', pathforbin);
            await download2file(ks.link, execpath);
            vscode.window.showInformationMessage("download Squash complete");
        }
    }
    // test after the download
    let exechash = await hash(execpath);
    // make sure its the one we expect:
    if (!hashesMatch(ks.checksum, exechash)) {
        // remove the bad binary.
        fs.unlinkSync(execpath);
        throw new Error("bad checksum for binary; download may be corrupted - please try again.");
    }
    fs.chmodSync(execpath, 0o755);
    return execpath;
}


// solo is the hash that was created from the squashctl binary when the binary was compiled
// gen is the hash that was generated locally from the squashctl file that the extension is trying to use
function hashesMatch(solo: string, gen: string): boolean {
    let hashParts = solo.split(" ");
    if (hashParts.length !== 2 || gen !== hashParts[0]) {
        return false;
    }
    return true;
}

function download2file(what: string, to: string): Promise<any> {

    return new Promise<any>((resolve, reject) => {
        let file = fs.createWriteStream(to);
        let stream = download(what);
        stream.pipe(file);
        file.on('close', resolve);
        file.on("finish", function () {
            file.close();
        });
        stream.on('error', reject);
        file.on('error', reject);

    });
}


function hash(f: string): Promise<string> {
    return new Promise<string>((resolve, reject) => {
        const input = fs.createReadStream(f);
        const hash = crypto.createHash('sha256');

        input.on('data', function (data: Buffer) {
            hash.update(data);
        });
        input.on('error', reject);
        input.on('end', () => {
            resolve(hash.digest("hex"));
        });

    });
}