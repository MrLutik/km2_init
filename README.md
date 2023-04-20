# KM
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