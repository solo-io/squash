'use strict';

import * as kube from './kube-interfaces';
import * as shelljs from 'shelljs';


import * as config from './config';
import * as cli from './cli';

import * as vscode from 'vscode';


/* Flow of this extension
   User opens extension
   Interactive selection of debug type, ns, pod, container
   Squashctl creates a debug connection, prints out needed information
   Extension parses the squashctl return value
   Extension uses vscode's debug capabilities
*/

/*
  Configuration values
  - remotePath - for source mapping, the source code path used when the target binary was compiled

*/



// this method is called when your extension is activated
// your extension is activated the very first time the command is executed
export function activate(context: vscode.ExtensionContext) {
    // Use the console to output diagnostic information (console.log) and errors (console.error)
    // This line of code will only be executed once when your extension is activated
    console.log(`TV5`);
    console.log(`Congratulations, your extension "Squash" is now active!`);

    let se = new SquashExtension(context);

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






export class DebuggerPickItem implements vscode.QuickPickItem {
    label: string;
    description: string;
    detail?: string;

    debugger: string;

    constructor(dbg: string) {
        this.label = `${dbg}`;
        this.description = dbg;
        this.debugger = dbg;
    }
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

export class NamespacePickItem implements vscode.QuickPickItem {
    label: string;
    description: string;
    detail?: string;

    namespace: kube.Namespace;

    constructor(namespace: kube.Namespace) {
        this.label = namespace.metadata.name;
        this.description = "namespace";
        this.namespace = namespace;
    }
}

export class ContainerPickItem implements vscode.QuickPickItem {
    label: string;
    description: string;
    detail?: string;

    container: kube.Container;

    constructor(container: kube.Container) {
        this.label = `${container.name} (${container.image})`
        this.description = "container";
        this.container = container;
    }
}
class SquashExtension {

    context: vscode.ExtensionContext;
    squashInfo: cli.SquashInfo;

    constructor(context: vscode.ExtensionContext) {
        this.context = context;
        this.squashInfo = cli.getSquashInfo();
    }

    async debug() {
        let squashpath: string = config.get_conf_or("path", null);
        if (!squashpath) {
            squashpath = await cli.getremote(this.context.extensionPath);
        }
        console.log("using squashctl from:");
        console.log(squashpath);
        if ( config.get_conf_or("verbose", false) ) {
            vscode.window.showInformationMessage("calling squashctl from: " + squashpath);
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
        //get namespace 
        let namespaces = await this.getNamespaces();

        let namespaceoptions: vscode.QuickPickOptions = {
            placeHolder: "Please select a namespace",
        };

        let namespaceItems: NamespacePickItem[] = namespaces.map(namespace => new NamespacePickItem(namespace));

        const ns_item = await vscode.window.showQuickPick(namespaceItems, namespaceoptions);

        if (!ns_item) {
            console.log("chosing namespace canceled - debugging canceled");
            return;
        }
        let selectedNamespace = ns_item.namespace;
        // get pod
        let pods = await this.getPods(selectedNamespace.metadata.name);

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

        // Get the specific Container.
        let containeroptions: vscode.QuickPickOptions = {
            placeHolder: "Please select a container",
        };

        let containers: kube.Container[] = this.getContainers(selectedPod);
        let containerItems: ContainerPickItem[] = containers.map(container => new ContainerPickItem(container));

        let selectedContainer: ContainerPickItem;
        if (containerItems.length === 1) {
            // If there is only one Container, automatically choose it.
            selectedContainer = containerItems[0];
        } else {
            const containerItem = await vscode.window.showQuickPick(containerItems, containeroptions);
            if (!containerItem) {
                console.log("choosing container canceled - debugging canceled");
                return;
            }
            selectedContainer = containerItem;
        }

        // choose debugger to use
        const debuggerList = ["dlv", "java"];
        let debuggerItems: DebuggerPickItem[] = debuggerList.map(name => new DebuggerPickItem(name));
        let debuggerOptions: vscode.QuickPickOptions = {
            placeHolder: "Please select a debugger",
        };
        const chosenDebugger = await vscode.window.showQuickPick(debuggerItems, debuggerOptions);
        if (!chosenDebugger) {
            console.log("chosing debugger canceled - debugging canceled");
            return;
        }
        console.log("You chose debugger: " + JSON.stringify(chosenDebugger));
        let debuggerName = chosenDebugger.debugger;
        let extraArgs = config.get_conf_or("extraArgs", "");

        let processMatch = config.get_conf_or("processMatch", "");

        // now invoke squashctl
        let cmdSpec = `${squashpath} ${extraArgs} --machine`;
        cmdSpec += ` --pod ${selectedPod.metadata.name}`;
        cmdSpec += ` --namespace ${selectedPod.metadata.namespace}`;
        cmdSpec += ` --container ${selectedContainer.container.name}`;
        cmdSpec += ` --debugger ${debuggerName}`;
        if (processMatch !== "") {
            cmdSpec += ` --process-match ${processMatch}`;            
        }
        console.log(`executing ${cmdSpec}`);
        let stdout = await exec(cmdSpec);
        let responseData = JSON.parse(stdout);
        if (!responseData) {
            throw new Error("can't parse output of squashctl: " + stdout);
        }

        let remotepath = config.get_conf_or("remotePath", null);

        // port forward
        let localport = await kubectl_portforward(responseData.PortForwardCmd);

        let localpath = workspace.uri.fsPath;
        // start debugging!
            let debuggerconfig;
            switch (debuggerName) {
                case "dlv":
                    debuggerconfig = {
                        name: "Remote",
                        type: "go",
                        request: "launch",
                        mode: "remote",
                        port: localport,
                        host: "127.0.0.1",
                        program: localpath,
                        remotePath: remotepath,
                        //      stopOnEntry: true,
                        env: {},
                        args: [],
                        showLog: true,
                        trace: "verbose"
                    };
                    break;
                case "java":
                    debuggerconfig = {
                        type: "java",
                        request: "attach",
                        name: "Attach to java process",
                        port: localport,
                        hostName: "127.0.0.1",
                    };
                    break;
                case "nodejs":
                case "nodejs8":
                    debuggerconfig = {
                        type: "node",
                        request: "attach",
                        name: "Attach to Remote",
                        address: "127.0.0.1",
                        port: localport,
                        localRoot: localpath,
                        remoteRoot: remotepath
                    };
                    break;
                case "python":
                    // TODO - add this to config when python enabled
                    let ptvsdsecret = config.get_conf_or("pythonSecret", "");
                    debuggerconfig = {
                        type: "python",
                        request: "attach",
                        name: "Python: Attach",
                        localRoot: localpath,
                        remoteRoot: remotepath,
                        port: localport,
                        secret: ptvsdsecret,
                        host: "127.0.0.1"
                    };
                    break;
                case "gdb":
                    let autorun: string[] = [];
                    if (remotepath) {
                        autorun = [`set substitute-path "${remotepath}" "${localpath}"`];
                    }
                    debuggerconfig = {
                        type: "gdb",
                        request: "attach",
                        name: "Attach to gdbserver",
                        target: "localhost:" + localport,
                        remote: true,
                        cwd: localpath,
                        autorun: autorun
                    };
                    break;
                default:
                    throw new Error(`Unknown debugger ${debuggerName}`);
            }

        return vscode.debug.startDebugging(
            workspace,
            debuggerconfig
        );

    }

    async  getPods(namespace: string): Promise<kube.Pod[]> {
        const podsjson = await kubectl_get<kube.PodList>("pods", "-n",namespace);
        return podsjson.items;
    }

    async  getNamespaces(): Promise<kube.Namespace[]> {
        const namespacesjson = await kubectl_get<kube.NamespaceList>("namespaces");
        return namespacesjson.items;
    }

    getContainers(pod: kube.Pod): kube.Container[] {
        return pod.spec.containers
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

function kubectl_portforward(cmd: string): Promise<number> {
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
        let child = shelljs.exec(maybeKubeEnv() + cmd, handler);
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

    console.log(["port forwarding on", JSON.stringify(p)]);
    return p;
}

function kubectl_get<T=any>(cmd: string, ...args: string[]): Promise<T> {
    return kubectl("get -o json " + cmd + " " + args.join(" ")).then(JSON.parse);
}

function kubectl(cmd: string): Promise<string> {
    return exec("kubectl" + maybeKubeConfig() + " " + cmd);
}

function maybeKubeConfig(): string {

    let maybeKubeConfig: string = config.get_conf_or("kubeConfig", null);
    if (!maybeKubeConfig) {
        maybeKubeConfig = "";
    } else {
        maybeKubeConfig = ` --kubeconfig="${maybeKubeConfig}" `;
    }
    return maybeKubeConfig;
}

function maybeKubeEnv(): string {

    let maybeKubeConfig: string = config.get_conf_or("kubeConfig", null);
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
        shelljs.exec(maybeKubeEnv() + cmd, options, handler);
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

