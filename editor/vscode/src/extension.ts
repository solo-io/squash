'use strict';

import * as kube from './kube-interfaces';
import * as shelljs from 'shelljs';

import * as fs from 'fs';
import * as path from 'path';
import * as download from 'download';
import * as crypto from 'crypto';

import squashVersionData = require('./squash.json');

// The module 'vscode' contains the VS Code extensibility API
// Import the module and reference it with the alias vscode in your code below
import * as vscode from 'vscode';

const OutPort = 1236;

const confname = "squash";

// this method is called when your extension is activated
// your extension is activated the very first time the command is executed
export function activate(context: vscode.ExtensionContext) {
    // Use the console to output diagnostic information (console.log) and errors (console.error)
    // This line of code will only be executed once when your extension is activated
    console.log(`Congratulations, your extension "${confname}" is now active!`);

    let se = new SquashExtention(context);

    // The command has been defined in the package.json file
    // Now provide the implementation of the command with  registerCommand
    // The commandId parameter must match the command field in package.json
    let disposable = vscode.commands.registerCommand('extension.debugPod', () => {
        // The code you place here will be executed every time your command is executed
        return se.debug().catch(handleError);
    });

    context.subscriptions.push(disposable);
}

// this method is called when your extension is deactivated
export function deactivate() { }

async function getremote(extPath: string): Promise<string> {
    let pathforbin = path.join(extPath, "binaries", getSquashInfo().version);
    let execpath = path.join(pathforbin, "kubsquash");

    let ks = getSquashctl();

    if (fs.existsSync(execpath)) {
        let exechash = await hash(execpath);
        // make sure its the one we expect:
        // this can happen on version updates.
        if (exechash !== ks.checksum) {
            // remove the bad binary.
            fs.unlinkSync(execpath);
        }
    }

    if (!fs.existsSync(execpath)) {
        let s = await vscode.window.showInformationMessage("HEY Download Squash?", "yes", "no");
        if (s === "yes") {
            vscode.window.showInformationMessage("download started");
            shelljs.mkdir('-p', pathforbin);
            await download2file(ks.link, execpath);
            vscode.window.showInformationMessage("download Squash complete");
        }
    }
    // TODO(mitchdraft) - renable
    // // test after the download
    // let exechash = await hash(execpath);
    // // make sure its the one we expect:
    // if (exechash !== ks.checksum) {
    //     // remove the bad binary.
    //     fs.unlinkSync(execpath);
    //     throw new Error("bad checksum for binary; download may be corrupted - please try again.");
    // }
    // fs.chmodSync(execpath, 0o755);
    return execpath;
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

export class PodPickItem implements vscode.QuickPickItem {
    label: string;
    description: string;
    detail?: string;

    pod: kube.Pod;

    constructor(pod: kube.Pod) {
        let podname = pod.metadata.name;
        let nodename = pod.spec.nodeName;
        this.label = `${podname} (${nodename})`;
        this.description = "pod";
        this.pod = pod;
    }
}
class SquashExtention {

    context: vscode.ExtensionContext;
    squashInfo: SquashInfo;

    constructor(context: vscode.ExtensionContext) {
        this.context = context;
        this.squashInfo = getSquashInfo();
    }

    async debug() {
        /*
            run the squashkube binary with -server
        */

        let squahspath: string = get_conf_or("path", null);
        if (!squahspath) {
            squahspath = await getremote(this.context.extensionPath);
        }

        if (!vscode.workspace.workspaceFolders) {
            throw new Error("no workspace folders");
        }

        let workspace: vscode.WorkspaceFolder;
        if (vscode.workspace.workspaceFolders.length === 0) {
            throw new Error("Can't start debugging without a project open");
        } else if (vscode.workspace.workspaceFolders.length === 1) {
            workspace = vscode.workspace.workspaceFolders[0];
        } else {
            let wfoptions: vscode.QuickPickOptions = {
                placeHolder: "Please a project to debug",
            };
            let wfItems = vscode.workspace.workspaceFolders.map(
                wf => new WorkspaceFolderPickItem(wf));

            const item = await vscode.window.showQuickPick(wfItems, wfoptions);

            if (item) {
                workspace = item.obj;
            } else {
                console.log("debugging canceled");
                return;
            }
        }

        /*
           get namespace and pod
        */
        let pods = await this.getPods();

        let podoptions: vscode.QuickPickOptions = {
            placeHolder: "Please select a pod",
        };

        let podItems: PodPickItem[] = pods.map(pod => new PodPickItem(pod));

        const item = await vscode.window.showQuickPick(podItems, podoptions);

        if (!item) {
            console.log("chosing pod canceled - debugging canceled");
            return;
        }
        let selectedPod = item.pod;

        let containerRepo = get_conf_or("containerRepository", null);
        let containerRepoArg = "";
        if (containerRepo) {
            containerRepoArg = `--container-repo ${containerRepo}`;
        }

        let extraArgs  = get_conf_or("extraArgs", "");
        // now invoke squashctl
        let stdout = await exec(maybeKubeEnv() + `${squahspath} ${extraArgs} ${containerRepoArg} --machine --debug-server --pod ${selectedPod.metadata.name} --namespace ${selectedPod.metadata.namespace} --debugger dlv`);
        let squashPodRegex = /pod.name:\s+(\S+)\s*$/g;
        let match = squashPodRegex.exec(stdout);
        if (match === null) {
            throw new Error("can't parse output of squashctl: " + stdout);
        }
        // get created pod name
        let squashPodName = match[1];
        // let pa = new PodAddress(selectedPod.metadata.namespace, squashPodName, OutPort);
        let pa = new PodAddress("squash-debugger", squashPodName, OutPort);

        let remotepath = get_conf_or("remotePath", null);

        // port forward
        let localport = await kubectl_portforward(pa);

        let localpath = workspace.uri.fsPath;
        // start debugging!
        let debuggerconfig: vscode.DebugConfiguration = {
            type: "go",
            name: "Remote",
            request: "launch",
            mode: "remote",
            port: localport,
            host: "127.0.0.1",
            program: localpath,
            remotePath: remotepath,
            //    stopOnEntry: true,
            env: {},
            args: [],
            showLog: true,
            trace: "verbose"
        };

        return vscode.debug.startDebugging(
            workspace,
            debuggerconfig
        );

    }

    async  getPods(): Promise<kube.Pod[]> {
        const podsjson = await kubectl_get<kube.PodList>("pods", "--all-namespaces");
        return podsjson.items;
    }

}

export class WorkspaceFolderPickItem implements vscode.QuickPickItem {
    label: string;
    description: string;
    detail?: string;
    obj: vscode.WorkspaceFolder;

    constructor(obj: vscode.WorkspaceFolder) {
        this.label = obj.name;
        this.obj = obj;
        this.description = "workspace";
    }
}

export class PodAddress {
    podName: string;
    podNamespace: string;
    port: number;

    constructor(podNamespace: string, podName: string, port: number) {
        this.podNamespace = podNamespace;
        this.podName = podName;
        this.port = port;
    }
}

function kubectl_portforward(remote: PodAddress): Promise<number> {

    let cmd = get_conf_or("kubectl-path", "kubectl") + maybeKubeConfig() + ` --namespace=${remote.podNamespace} port-forward ${remote.podName} :${remote.port}`;
    console.log("Executing: " + cmd);
    let p = new Promise<number>((resolve, reject) => {
        let resolved = false;
        let handler = function (code: number, stdout: string, stderr: string) {
            if (resolved !== true) {
                if (code !== 0) {
                    reject(new ExecError(code, stdout, stderr));
                } else {
                    reject(new Error("Didn't receive port"));
                }
            } else {
                console.log(`port forward ended unexpectly: ${code} ${stdout} ${stderr} `);
            }
        };
        let child = shelljs.exec(cmd, handler);
        let stdout = "";
        child.stdout.on('data', function (data) {
            stdout += data;
            let portRegexp = /from\s+.+:(\d+)\s+->/g;
            let match = portRegexp.exec(stdout);
            if (match !== null) {
                resolved = true;
                resolve(parseInt(match[1]));
            }
        });
    });

    return p;
}

function kubectl_get<T=any>(cmd: string, ...args: string[]): Promise<T> {
    return kubectl("get -o json " + cmd + " " + args.join(" ")).then(JSON.parse);
}

function kubectl(cmd: string): Promise<string> {
    return exec(get_conf_or("kubectl-path", "kubectl") + maybeKubeConfig() + " " + cmd);
}

function maybeKubeConfig(): string {

    let maybeKubeConfig: string = get_conf_or("kubeConfig", null);
    if (!maybeKubeConfig) {
        maybeKubeConfig = "";
    } else {
        maybeKubeConfig = ` --kubeconfig="${maybeKubeConfig}" `;
    }
    return maybeKubeConfig;
}

function maybeKubeEnv(): string {

    let maybeKubeConfig: string = get_conf_or("kubeConfig", null);
    if (!maybeKubeConfig) {
        maybeKubeConfig = "";
    } else {
        maybeKubeConfig = `KUBECONFIG="${maybeKubeConfig}" `;
    }
    return maybeKubeConfig;
}

// https://github.com/Microsoft/TypeScript/wiki/Breaking-Changes#extending-built-ins-like-error-array-and-map-may-no-longer-work
class ExecError extends Error {
    code: number;
    stderr: string;
    stdout: string;

    constructor(code: number, stdout: string, stderr: string) {
        super((stdout + stderr).trim());

        // Set the prototype explicitly.
        Object.setPrototypeOf(this, ExecError.prototype);

        this.code = code;
        this.stderr = stderr;
        this.stdout = stdout;
    }
}

async function exec(cmd: string): Promise<string> {
    console.log("Executing: " + cmd);
    let promise = new Promise<string>((resolve, reject) => {
        let handler = function (code: number, stdout: string, stderr: string) {
            if (code !== 0) {
                reject(new ExecError(code, stdout, stderr));
            } else {
                resolve(stdout);
            }
        };

        let options = {
            async: true,
            stdio: ['ignore', 'pipe', 'pipe'],
        };
        shelljs.exec(cmd, options, handler);
    });

    return promise;
}

const handleError = (err: Error) => {
    if (err) {
        if (err.message) {
            vscode.window.showErrorMessage(err.message);
        } else {
            vscode.window.showErrorMessage("Unknown error has occurred");
        }
    }
};

function get_conf_or(k: string, d: any): any {
    let config = vscode.workspace.getConfiguration(confname);
    let v = config[k];
    if (!v) {
        return d;
    }
    return v;
}

class BinariesSha {
    linux!: string;
    darwin!: string;
}
class SquashInfo {
    version!: string;
    baseName!: string;
    binaries!: BinariesSha;
}

function getSquashInfo(): SquashInfo {
    return <SquashInfo>squashVersionData;
}

interface SquashctlBinary {
    link: string;
    checksum: string;
}

function createSquashctlBinary(os: string, checksum: string): SquashctlBinary {
    return {
        link: "https://github.com/solo-io/squash/releases/download/" + getSquashInfo().version + "/" + getSquashInfo().baseName + "-" + os,
        checksum: checksum
    };
}

function getSquashctl(): SquashctlBinary {
    // download the squash version for this extension
    var osver = process.platform;
    switch (osver) {
        case 'linux':
            return createSquashctlBinary("linux", getSquashInfo().binaries.linux);
        case 'darwin':
            return createSquashctlBinary("osx", getSquashInfo().binaries.darwin);
        default:
            throw new Error(osver + " is current unsupported");
    }
}
