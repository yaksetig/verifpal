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
The model below demonstrates the symbolic `PedersenCommit` and `Neg` primitives. It shows that adding a commitment to its negation simplifies to zero and checks that the committed value remains secret from a passive attacker. Use the `GROUPADD`, `Neg`, and `SCALARNEG` primitives to express group arithmetic; Verifpal's core syntax does not include infix `+` or `-` operators.

File: `examples/pedersen_commit_demo.vp`
```verifpal
attacker[passive]

principal Alice[
    knows private v
    knows private r
    Commit = PedersenCommit(v, r)
    Cancel = GROUPADD(Commit, Neg(Commit))
    DoubleCancel = GROUPADD(Cancel, GROUPADD(Commit, Neg(Commit)))
]

queries[
    equivalence? Cancel, 0
    equivalence? DoubleCancel, 0
    confidentiality? v
]
```

Run the model with:
```sh
./build/verifpal verify examples/pedersen_commit_demo.vp
```
Successful verification confirms that the commitment arithmetic simplifies correctly and that `v` is kept secret in the passive attacker model.

## Example: Zero-Knowledge Proof Primitives
The `ZKSETUP`, `ZKPROVE`, and `ZKVERIFY` primitives let you model abstract zero-knowledge protocols:
- `ZKSETUP(seed)` creates reusable public parameters from a seed.
- `ZKPROVE(params, statement, witness)` produces a proof that the `witness` satisfies the given `statement` under the published
  `params`.
- `ZKVERIFY(params, statement, proof)` acts like an assertion: it succeeds only when the proof was produced for the same
  parameters *and* statement, and otherwise fails immediately without needing a follow-up equality check.

File: `examples/zkproof_demo.vp`
```verifpal
attacker[active]

principal Setup[
    generates setup_seed
    Params = ZKSETUP(setup_seed)
]

Setup -> Alice: Params
Setup -> Bob: Params

principal Alice[
    knows private secret
    Statement = HASH(secret)
    Proof = ZKPROVE(Params, Statement, secret)
]

Alice -> Bob: Statement, Proof

principal Bob[
    # Acts as a guard: verification fails here if the proof was forged or mismatched.
    ZKVERIFY(Params, Statement, Proof)
]

queries[
    confidentiality? secret
]
```

Run the model with:
```sh
./build/verifpal verify examples/zkproof_demo.vp
```
Verification succeeds when the proof and statement match, enforcing the assertion at Bob’s verification step while keeping
the witness `secret` confidential.
