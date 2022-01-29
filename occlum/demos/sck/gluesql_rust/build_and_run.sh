#!/bin/bash
set -e

# compile rust_app
pushd gluesql-test
occlum-cargo build
popd

# initialize occlum workspace
rm -rf occlum_instance && mkdir occlum_instance && cd occlum_instance

occlum init && rm -rf image
copy_bom -f ../rust-demo.yaml --root image --include-dir /opt/occlum/etc/template

sed -i '3 s/32/128/' Occlum.json
occlum build
occlum run /bin/gluesql-test
