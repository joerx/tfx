# TFX

Razor-thin wrapper around Terraform that will download & install the latest matching version for the Terraform version constraint in the current module.

Releases are parsed and downloaded from the official releases page at https://releases.hashicorp.com/terraform/

## Usage

Given some terraform config:

```terraform
terraform {
    required_version = "~> 0.12.0"
}
```

Running `tfx` will download the latest official version on the `0.12.0` branch and run it. Any arguments are directly forwarded to the Terraform binary.

Example:

```
$ tfx version
2020/10/01 15:46:22 Trying to find terraform version matching "~> 0.12.0"
2020/10/01 15:46:22 Using version 0.12.29
2020/10/01 15:46:22 Exec '/tmp/tfversions/0.12.29/terraform version'
Terraform v0.12.29
```

## Limitations

This is an extremely simple PoC and might crash and burn in more complicated scenarios. For example:

- It doesn't have any fallback if no matching terraform version can be found
- Undefined behaviour of more or less than exactly 1 `required_version` constraint exists
- There are no tests
