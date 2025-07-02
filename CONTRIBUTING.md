# Contributing to \[Schema]

Thank you for considering contributing to this Go library! We welcome contributions of all kinds—bug reports, feature requests, documentation updates, tests, and code improvements.

Please take a moment to review this guide before submitting an issue or pull request.

## 🚀 Getting Started

1. **Fork** the repository.

2. **Clone** your fork:

   ```bash
   git clone https://github.com/your-username/your-library.git
   cd your-library
   ```

3. **Install dependencies**:

   ```bash
   go mod tidy
   ```

4. **Run tests** to ensure everything is working:

   ```bash
   go test ./...
   ```

## 🧑‍💻 Development Guidelines

* Go version: \[Specify Go version, e.g., `1.23+`]
* Use idiomatic Go and follow the official [Effective Go](https://golang.org/doc/effective_go.html) guide.
* All code should be tested. We use [Go's testing package](https://golang.org/pkg/testing/) and optionally [Testify](https://github.com/stretchr/testify).
* Keep the API stable and backward compatible when possible.

## 🤝 How to Contribute

### 🐛 Reporting Bugs

* Use the [issue tracker](https://github.com/your-org/your-library/issues) to report bugs.
* Include steps to reproduce, expected behavior, and actual behavior.

### 💡 Suggesting Enhancements

* Propose enhancements via issues before working on a large feature.
* Provide context, use cases, and possible implementation ideas.

### ✅ Submitting a Pull Request

1. Create a new branch from `main`:

   ```bash
   git checkout -b feat/your-feature-name
   ```

2. Make your changes.

3. Run tests and format your code:

   ```bash
   go fmt ./...
   go test ./...
   ```

4. Commit with a clear message:

   ```bash
   git commit -m "feat: add support for XYZ"
   ```

5. Push and open a PR against the `main` branch.

6. Follow the PR template and reference any related issues.

## 🧼 Code Style

* Run `go fmt ./...` before submitting.
* Use descriptive variable and function names.
* Use `golangci-lint` or similar tools to catch linting issues.

## 🧪 Running Tests

Use the standard testing command:

```bash
go test -v ./...
```

To run with coverage:

```bash
go test -cover ./...
```

## 📄 License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE) of the project.

Thank you again for your support! 🙌