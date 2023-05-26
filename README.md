# About

This is a [WSO2IS (WSO2 Identity Server)](https://wso2.com/identity-server/) playground.

This is somewhat related to https://github.com/rgl/terraform-keycloak.

# Usage

Install docker compose.

Add the following entries to your machine `hosts` file:

```
127.0.0.1 wso2is.test
127.0.0.1 mail.test
127.0.0.1 example-go-confidential.test
127.0.0.1 example-go-confidential-keycloak-saml-federation.test
```

Start the environment:

```bash
./create.sh
```

When anything goes wrong, you can try to troubleshoot at:

* `docker compose logs --follow`
* https://wso2is.test:9443/console/ (WSO2IS Console; login as `admin`:`admin`)
* https://wso2is.test:9443/carbon/ (WSO2IS Carbon Management Console; login as `admin`:`admin`)
* https://wso2is.test:9443/myaccount/ (WSO2IS My Account; login as `admin`:`admin`)
* https://wso2is.test:9443/oauth2/token/.well-known/openid-configuration (WSO2IS IdP OIDC metadata/configuration)
* https://wso2is.test:9443/identity/metadata/saml2 (WSO2IS IdP SAML metadata/configuration)
* http://keycloak.test:8080/realms/example/.well-known/openid-configuration (Keycloak IdP OIDC metadata/configuration)
* http://keycloak.test:8080/realms/example/protocol/saml/descriptor (Keycloak IdP SAML metadata/configuration)
* http://keycloak.test:8080 (Keycloak; login as `admin`:`admin`)
* http://mail.test:8025 (MailHog (email server))
* For SAML troubleshooting, you can use the browser developer tools to capture
  the requests/responses and paste them in the SAML Decoder & Parser at
  https://www.scottbrady91.com/tools/saml-parser.

Destroy everything:

```bash
./destroy.sh
```

# References

* https://is.docs.wso2.com
  * https://is.docs.wso2.com/en/latest/references/concepts/authentication/discovery/
* https://github.com/wso2/product-is
* https://github.com/wso2/docker-is
* https://www.scottbrady91.com/tools/saml-parser
