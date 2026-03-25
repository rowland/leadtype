# Samples

The sample programs are now exposed through a single runner in this directory.

List the available samples:

```bash
go run ./samples -list
```

Run a sample and write its PDF into `samples/`:

```bash
go run ./samples test_003_hello_world
```

Open the generated PDF after writing it:

```bash
go run ./samples -o test_003_hello_world
```

The `-o` and `-open` flags both enable opening the generated file with the
platform `open` command.

Samples `test_003_hello_world`, `test_004_afm_fonts`, and
`test_007_i18n_afm` keep their historical Firefox preference when opened with
`-o`, since Preview does not handle their Type1 font output well.
