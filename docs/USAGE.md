# Verifpal Repository Usage

This guide summarizes how to build and run Verifpal from this repository and provides a ready-made model that exercises the new `PedersenCommit` primitive.

## Prerequisites
- Go toolchain installed (required for building). For installation help, follow the [official Go instructions](https://go.dev/doc/install).

## Build from Source
1. Install Verifpal's Go dependencies:
   ```sh
   make dep
   ```
2. Build the cross-platform binaries:
   ```sh
   make all
   ```
   Compiled binaries will be available under `build/`.

## Running the CLI
The Verifpal CLI provides several subcommands:
- `verify [model.vp]`: analyze a Verifpal model.
- `translate coq [model.vp]`: generate a Coq template.
- `translate pv [model.vp]`: generate a ProVerif template.
- `pretty [model.vp]`: pretty-print a model.

After building, run commands using the binary in `build/` (or `verifpal` if installed globally). For example:
```sh
./build/verifpal verify examples/pedersen_commit_demo.vp
```

## Example: Testing `PedersenCommit`
The model below demonstrates the symbolic `PedersenCommit` and `Neg` primitives. It shows that adding a commitment to its negation simplifies to zero and checks that the committed value remains secret from a passive attacker.

File: `examples/pedersen_commit_demo.vp`
```verifpal
attacker[passive]

principal Alice[
    knows private v
    knows private r
    Commit = PedersenCommit(v, r)
    Cancel = PedersenCommit(v, r) + PedersenCommit(-v, -r)
    DoubleCancel = Cancel + PedersenCommit(v, r) + PedersenCommit(-v, -r)
]

queries[
    equivalence? Cancel, 0
    equivalence? DoubleCancel, 0
    secrecy? v
]
```

Run the model with:
```sh
./build/verifpal verify examples/pedersen_commit_demo.vp
```
Successful verification confirms that the commitment arithmetic simplifies correctly and that `v` is kept secret in the passive attacker model.
