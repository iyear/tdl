---
title: "Extensions 🆕"
weight: 70
---

# Extensions

{{< hint warning >}}
Extensions are a new feature in tdl. They are still in the experimental stage, and the CLI may change in future versions.

If you encounter any problems or have any suggestions, please [open an issue](https://github.com/iyear/tdl/issues/new/choose) on GitHub.
{{< /hint >}}

## Overview

tdl extensions are add-on tools that integrate seamlessly with tdl. They provide a way to extend the core feature set of tdl, but without requiring every new feature to be added to the core.

tdl extensions have the following features:

- They can be added and removed without impacting the core tdl tool.
- They integrate with tdl, and will show up in tdl help and other places.

tdl extensions live in `~/.tdl/extensions`, which is controlled by `tdl extension` commands.

To get started with extensions, you can use the following commands:

{{< command >}}
tdl extension install iyear/tdl-whoami
{{< /command >}}

{{< command >}}
tdl whoami
{{< /command >}}

You can see the output of the `tdl-whoami` extension. Refer to the [tdl-whoami](https://github.com/iyear/tdl-whoami) for details.
```
You are XXXXX. ID: XXXXXXXX
```

## Finding extensions

You can find extensions by browsing [repositories with the `tdl-extension` topic](https://github.com/topics/tdl-extension).

## Installing extensions

To install an extension, use the `extension install` subcommand.

There are two types of extensions:

- `GitHub` : Extensions hosted on GitHub repositories.

    {{< command >}}
    tdl extension install <owner>/<repo>
    {{< /command >}}

    To install an extension from a private repository, you must set up a [GitHub personal access token](https://github.com/settings/personal-access-tokens/new)(with `Contents` read permission) in your environment with the `GITHUB_TOKEN` variable.

    {{< command >}}
    export GITHUB_TOKEN=YOUR_TOKEN
    tdl extension install <owner>/<private-repo>
    {{< /command >}}

- `Local` : Extensions stored on your local machine.
    
    {{< command >}}
    tdl extension install /path/to/extension
    {{< /command >}}

To install an extension even if it exists, use the `--force` flag:

{{< command >}}
tdl extension install --force EXTENSION
{{< /command >}}

To install multiple extensions at once, use the following command:

{{< command >}}
tdl extension install <owner>/<repo1> /path/to/extension2 ...
{{< /command >}}

To only print information without actually installing the extension, use the `--dry-run` flag:

{{< command >}}
tdl extension install --dry-run EXTENSION
{{< /command >}}

If you already have an extension by the same name installed, the command will fail. For example, if you have installed `foo/tdl-whoami`, you must uninstall it before installing `bar/tdl-whoami`.

## Running extensions

When you have installed an extension, you run the extension as you would run a native tdl command, using `tdl EXTENSION-NAME`. The `EXTENSION-NAME` is the name of the repository that contains the extension, minus the `tdl-` prefix.

For example, if you installed the extension from the `iyear/tdl-whoami` repository, you would run the extension with the following command.

{{< command >}}
tdl whoami
{{< /command >}}

Global config flags are still available when running an extension. For example, you can run the following command to specify namespace and proxy when running the `tdl-whoami` extension.

{{< command >}}
tdl -n foo --proxy socks5://localhost:1080 whoami
{{< /command >}}

Flags specific to an extension can also be used. For example, you can run the following command to enable verbose mode when running the `tdl-whoami` extension.

{{< hint info >}}
Remember to write global flags before extension subcommands and write extension flags after extension subcommands:
{{< command >}}
tdl <global-config-flags> <extension-name> <extension-flags>
{{< /command >}}

{{< /hint >}}

{{< command >}}
tdl -n foo whoami -v
{{< /command >}}

You can usually find specific information about how to use an extension in the README of the repository that contains the extension.

## Viewing installed extensions

To view all installed extensions, use the `extension list` subcommand. This command will list all installed extensions, along with their authors and versions.

{{< command >}}
tdl extension list
{{< /command >}}

## Updating extensions

To update an extension, use the `extension upgrade` subcommand. Replace the `EXTENSION` parameters with the name of extensions.

{{< command >}}
tdl extension upgrade EXTENSION1 EXTENSION2 ...
{{< /command >}}

To update all installed extensions, keep the `EXTENSION` parameter empty.

{{< command >}}
tdl extension upgrade
{{< /command >}}

To upgrade an extension from a GitHub private repository, you must set up a [GitHub personal access token](https://github.com/settings/personal-access-tokens/new)(with `Contents` read permission) in your environment with the `GITHUB_TOKEN` variable.

{{< command >}}
export GITHUB_TOKEN=YOUR_TOKEN
tdl extension upgrade EXTENSION
{{< /command >}}

To only print information without actually upgrading the extension, use the `--dry-run` flag:

{{< command >}}
tdl extension upgrade --dry-run EXTENSION
{{< /command >}}

## Uninstalling extensions

To uninstall an extension, use the `extension remove` subcommand. Replace the `EXTENSION` parameters with the name of extensions.

{{< command >}}
tdl extension remove EXTENSION1 EXTENSION2 ...
{{< /command >}}

To only print information without actually uninstalling the extension, use the `--dry-run` flag:

{{< command >}}
tdl extension remove --dry-run EXTENSION
{{< /command >}}

## Developing extensions

Please refer to the [tdl-extension-template](https://github.com/iyear/tdl-extension-template) repository for instructions on how to create, build, and publish extensions for tdl.
