#!/bin/sh -e
cd editor/vscode
vsce publish -p $VSCODE_TOKEN
