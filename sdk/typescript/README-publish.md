# Publishing the TypeScript SDK

```bash
npm run build
npm publish --access public
```

The package is `@axisrobo/ontovela` (Apache-2.0). `npm test` runs the contract
suite against the compiled output; the CI `typescript` job verifies before
release. Add a version bump to `package.json` for each release tag.
