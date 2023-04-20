# KM
This is a simple application to launch Docker containers using the specified base image and options. It supports the latest version of base-image and can launch containers with sekai and interx options.

## Usage

```
go run main.go [options]
```
## Options
- -image: Base-image version. Default: v0.13.7.
- -sekai: Set to true to start a container with sekai.
- -interx: Set to true to start a container with interx.

## Example
To run the application with the base-image version v0.13.7 and start a container with sekai:

```
go run main.go -image v0.13.7 -sekai
```
## Functionality
The application does the following:

1. It parses command-line flags to get the base-image version and container options.
2. It validates the provided base-image version against a SemVer format.
3. It constructs the base-image name using the provided version.
4. It verifies the image signature using cosign and checks if the image is verified.
5. It pulls the Docker image with the specified base-image name.

### Functions

#### Func VerifyImageSignature
1. Decode the PEM-encoded public key data.
2. Parse the public key from the decoded PEM block.
3. Assert that the parsed public key is of type *ecdsa.PublicKey.
4. Parse the `image reference`*.
5. Load the ECDSA verifier with the parsed public key.
6. Create a cosign.CheckOpts instance with the verifier.
7. Call cosign.VerifyImageSignatures with the context, `image reference`, and check options.
8. Returns boolean (true if verified) and error

* NB! `image reference` 
```
[<registry>/][<namespace>/]<repository>[:<tag>]
```