import * as vscode from 'vscode';

// this is the key in the vscode config map for which our squash configuration object is the value
const confname = "squash";

export function get_conf_or(k: string, d: any): any {
    let config = vscode.workspace.getConfiguration(confname);
    let v = config[k];
    if (!v) {
        return d;
    }
    return v;
}