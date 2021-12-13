# AWS EC2 WRAPPER

An opinionated way launch an EC2 instance and manage its lifecycle. This is used to test and reproduce certain
scenarios.

## 1. Params  File

A file with YAML format must be created at `~/.aws/params` with the following data. This is for externalization and to
avoid typing the params each time.

```yaml
#YAML File
ec2:
  instanceType: m5.large
  # Name of the keypair to ssh into the machine for launching the ec2 instance
  sshKeyPairName: keypair
  # Pre-existing security group for launching the ec2 instance
  securityGroupId: sg-00f****
  # Pre-existing subnet id for launching the ec2 instance   
  subNetId: subnet-****
  # This is to filter the AMIs that will be used to launch. 
  # The AMI with this value for the `Type` tag only is selected 
  amiTypeTag: my_template
  # This value will be added to the `Type` tag of the newly launched EC2 instance
  instanceTypeTag: my_template
  # This value will be the prefix of the `Name` tag of new EC2 Instance
  # This will be suffixed by the param set in the launch command
  instanceNameTagPrefix: my_inst_
```

## 2. AWS Credentials

The `~/.aws/params` file must be created for the corresponding `~/.aws/credentials` file. This can be managed with
the `awss` tool

## 3. `AWX` tool

The commands are

- `awx ls`
- `awx ami ls`
- `awx launch <amiId> <namePrefix>`
- `awx start <instanceId>`
- `awx stop <instanceId>`
- `awx terminate <instanceId>`

### 3.1 `awx ls`

Lists the EC2 Instances which matches the tag `Type` eq  `$instanceTypeTag`_(from param file)_. An optional command line
param `--raw` can be added see the output in json.

### 3.2 `awx ami ls`

Lists the AMI which matches the tag `Type` eq  `$atom_template`_(from param file)_. An optional command line
param `--raw` can be added see the output in json.

### 3.3 `awx launch <amiId> <namePrefix>`

Launches an EC2 Instance with the AMI with the `amiId`. The other params from the param file will be set as the
arguments while launching. The `Type` tag of the launched will be set with the value of `$atom_template`. The `Name` tag
will be created as `$instanceNameTagPrefix$namePrefix`.

The params from the `paramFile` can be overridden by command line args.

```
awx launch <amiId> <namePrefix> --ec2-instance-type m5.xlarge \
    --ec2-instance-name-tag-prefix my_inst
```

### 3.4 `awx start <instanceId>`

Starts an existing Ec2 Instance

### 3.5 `awx stop <instanceId>`

Starts an existing Ec2 Instance. This will ask for a user prompt

### 3.6 `awx terminate <instanceId>`

Terminates an existing Ec2 Instance. This will ask for a user prompt

## 4. AWS Context Switcher `AWSS`

This is a context switcher for the files in the directory `~/.aws`. The files must be named in the following format i.e.
ends with `.$context`.

```
credentials.dev
config.dev
params.dev

credentials.prod
config.prod
params.prod
```

In this case the context is `dev` and `prod`. We can switch the context to `dev` by running the command `awss dev`. This
will create COPY as shown below. The original files will be preserved

```
credentials.dev -> credentials
config.dev -> config
params.dev -> params
```

Now the `aws` commandline tool will use the correct files

### 4.1 Commands

- `awss` - list the contexts
- `awss <context>` activates the context
