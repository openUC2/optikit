# optikit

A system of record for product data management at openUC2

This repo is a prototype of [DN 10: Optikit platform](https://app.notion.com/p/DN-10-Optikit-platform-2854e612c78a800cb8e3c3ca7b5a3e76?source=copy_link).

## Usage

### Development

These instructions are for being able to run all the CI checks locally on your computer:

Prerequisites:

- [go](https://go.dev/doc/install)
- [uv](https://docs.astral.sh/uv/getting-started/installation/) (for manipulating STEP files and generating NewSwitch configs)
- [npm](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm) (for checking GLTF files)

Then run:

```bash
make install
```

Now you can run the CI checks:
```bash
make generate
make test
```

### Generating NewSwitch configs

#### Download/install `optikit`

First, you will need to download the `optikit` tool, which is available as a single self-contained
executable file. You should visit this repository's
[releases page](https://github.com/openUC2/optikit/releases/latest) and download an archive
file for your platform and CPU architecture; for example, on a Raspberry Pi 4, you should probably
download the archive named `optikit_{version number}_linux_arm64.tar.gz` (where the version number
should be substituted). You can extract the `optikit` binary from the archive using a command like:
```bash
tar -xzf optikit_{version number}_{os}_{cpu architecture}.tar.gz optikit
```

Then you may need to move the `optikit` binary into a directory in your system path, or you can just
run the `optikit` binary by its absolute/relative path (in which case you should replace `optikit`
with the absolute/relative path of the binary in the example commands listed below).

#### Clone the optikit repo

Next, you will need to clone this repo to your computer, e.g.:

```bash
git clone https://github.com/openUC2/optikit.git
```

This will serve two purposes:
1. It provides an example DSN model you can use for generating a NewSwitch config
2. It provides a Python script which is used as part of the NewSwitch config-generation process

Now we will install the dependencies needed for the Python script. In the repo, run:

```bash
uv sync
```

#### Generate an assembly report

Now we will generate an assembly report from a DSN model. For this example, we will use
[examples/designs/microscopes/fluorescence.dsn](./examples/designs/microscopes/fluorescence.dsn/),
whose NewSwitch config components are in the `wiring` assembly.
Run:

```bash
cd examples/designs/microscopes/fluorescence.dsn
optikit dev dsn comp report-assm --assembly wiring your-output-assembly-report-file.yml
```

This will create the file `your-output-assembly-report-file.yml`.

#### Generate a NewSwitch config

Now we will generate a NewSwitch config from the assembly report. Still in the `fluorescence.dsn`
directory, run:

```bash
uv run poe config your-output-assembly-report-file.yml your-output-newswitch-config.yml
```

This will create the file `your-output-newswitch-config.yml`, which is the NewSwitch config.

## Licensing

We have chosen the following licenses in order to give away our work for free, so that you can
freely use it for whatever purposes you have, with minimal restrictions, while still protecting our
disclaimer that this work is provided without any warranties at all. If you're using this project,
or if you have questions about the licenses, we'd love to hear from you - please start a new
discussion thread in the "Discussions" tab of this repository on Github or email us at
<lietk12@gmail.com> .

### Software

Except where otherwise indicated, source code provided here is covered by the following information:

**Copyright Ethan Li and Optikit project contributors**

SPDX-License-Identifier: `Apache-2.0 OR BlueOak-1.0.0`

Software files in this repository are released under the
[Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0) and the
[Blue Oak Model License 1.0.0](https://blueoakcouncil.org/license/1.0.0);
you can use the source code provided here either under the Apache License or under the
Blue Oak Model License, and you get to decide which license you will agree to.
We are making the software available under the Apache license because it's
[OSI-approved](https://writing.kemitchell.com/2019/05/05/Rely-on-OSI.html),
but we like the Blue Oak Model License more because it's easier to read and understand.
Please read and understand the licenses for the specific language governing permissions and
limitations.

### Everything else

Except where otherwise indicated in this repository, any other files (such as images, media, data,
and textual documentation) provided here not already covered by the software license (described
above) are instead covered by the following information:

**Copyright Ethan Li and Optikit project contributors**

SPDX-License-Identifier: `CC-BY-4.0`

Files in this project are released under the
[Creative Commons Attribution 4.0 International License](http://creativecommons.org/licenses/by/4.0/).
Please read and understand the license for the specific language governing permissions and
limitations.
