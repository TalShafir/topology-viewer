#!/bin/bash

TAG=v0.0.1

jq '[.[] | select(.type=="Archive") | 
{
    bin: .extra.Binaries[0],
    sha256: .extra.Checksum | sub("^sha256:"; ""),
    uri:("https://github.com/TalShafir/topology-viewer/releases/download/'$TAG'/"+.name),
    selector: {
        matchLabels: {
            os: .goos,
            arch: .goarch
        }
    }
}]' dist/artifacts.json | yq -P -o yaml - | \
yq ea -i "select(fileIndex == 0).spec.platforms *= select(fileIndex == 1) | select(fileIndex == 0) | .spec.version=\"$TAG\"" topology-viewer.yaml -