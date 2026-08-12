# Publishing the Python SDK

```bash
python -m build
python -m twine upload dist/*
```

The distribution is `ontovela` (Apache-2.0). `python -m unittest
tests.test_client -v` verifies before release. Add a version bump to
`pyproject.toml` for each release tag.
