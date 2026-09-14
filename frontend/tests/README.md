# Project browser regression

Start a local frontend built from the current source (an isolated checkout is
recommended when another dev server is running), then run:

```sh
TEST_BASE_URL=http://127.0.0.1:3317 node tests/projects-regression.cjs
```

Requires the project's Playwright dependency and an installed Chromium browser.
Alternatively, set `TEST_CHROME_PATH` to an installed Chrome executable.
All API requests are fulfilled with local fixtures; no actual login or database
writes occur. Checks cover project list/detail rendering, form opening/cancel,
invalid task rejection, successful task/budget payloads, query refresh, mobile
form interaction, and browser runtime errors. This complements TypeScript/lint
and backend contract/authorization tests; it does not replace live staging UAT.
