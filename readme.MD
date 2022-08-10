# Usage
1. Create account and get license in web https://www.maxmind.com/en/accounts/current/people/current?utm_source=txn_email&utm_medium=email&utm_campaign=account
2. Build docker image
   `docker build . -t myapp`

3. Run the image with following environment variables:
```
CRON_EXPRESSION=0 9 * * 3
LICENSE=ZGvuxCAnfiIe1E7l
```
4. Run above image in traefik yaml.