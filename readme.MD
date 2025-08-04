# SecureHTTP

HTTP server extended wrapper based on Golang Fiber ([references](https://docs.gofiber.io/)). It supports multiple out of the box implementations such as:

1. Gracefull shutdown
2. Defer close function
3. Pre-installed middleware
4. Powerful structured logging
5. Default path for kubernetes Probe check
6. Profiling API
7. Open API Specs setup in yaml source
8. Status API to maintain released version
9. JSON API Specification response writter ([references](https://jsonapi.org/))

## Middlewares

Here are the list of pre-installed middleware out of the box:

1. Panic recovery
2. RequestID injector
3. Logging
4. Advanced HTTP info constructor
5. Security headers
6. CORS Setup
7. Rate limit
8. Open telemetry

## How to?

See example directory for more practical setup.

## Contribution

1. Create new pull request
2. Please follow conventional commit standard ([references](https://www.conventionalcommits.org/en/v1.0.0/))
3. There will be new release after PR merged
4. So, that's why it should follow conventional commit

## Release New Version

1. Pull latest from main
2. Use `npx commit-and-tag-version` then push the along with the tag ([references](https://www.npmjs.com/package/commit-and-tag-version/))
