#!/bin/sh -e
cd editor/vscode
npm install -g typescript # TODO - put in cloudbuilder
vsce publish -p $VSCODE_TOKEN
