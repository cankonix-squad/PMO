# Feature boundaries

Feature code belongs under its bounded context (`projects`, `dashboard`,
`reports`, `governance`, `users`, or `integrations`). Shared UI remains under
`components`; feature services and schemas must not be imported by unrelated
features without an explicit shared contract.
