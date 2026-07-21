# Contributing to the Open Knowledge Standard (OKF)

We welcome contributions from the open-source community!

## Getting Started

1. **Fork & Clone**: Fork the repository and clone it locally.
2. **Setup Virtual Environment**:
   ```bash
   python3 -m venv .venv
   source .venv/bin/activate
   pip install -e okf
   ```
3. **Develop & Test**: Make your changes and run the test suite:
   ```bash
   pytest okf/tests
   ```
4. **Submit a Pull Request**: Submit a clear PR describing your improvements or bug fixes.

## Code Style & Testing
- Ensure Python code follows standard formatting and type hints.
- Add unit tests under `okf/tests/` for new features or extractors.
