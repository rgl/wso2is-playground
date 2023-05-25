# see https://github.com/wso2/product-is/blob/master/oidc-conformance-tests/configure_is.py
# see https://github.com/wso2/docker-is/blob/master/docker-compose/is/docker-compose.yml

from base64 import b64encode
from pathlib import Path
from urllib.parse import urlparse, urljoin, quote
import json
import logging
import requests
import time


def wait_for_ready(base_url):
    while True:
        try:
            r = requests.get(f'{base_url}/carbon/admin/login.jsp', verify=False)
            if r.status_code == 200:
                break
        except:
            time.sleep(3)


def find_application(base_url, name):
    headers = {
        'Authorization': f'Basic {b64encode("admin:admin".encode("utf-8")).decode("ascii")}',
        'Content-Type': 'application/json',
    }
    applications_url = f'{base_url}/api/server/v1/applications'
    resp = requests.get(
        url=applications_url,
        params={
            'filter': f'name eq {name}',
        },
        headers=headers,
        verify=False)
    resp.raise_for_status()
    response = resp.json()
    if response['count'] == 0:
        return None
    return response['applications'][0]


def get_application(base_url, name):
    application = find_application(base_url, name)
    if not application:
        return None
    headers = {
        'Authorization': f'Basic {b64encode("admin:admin".encode("utf-8")).decode("ascii")}',
        'Content-Type': 'application/json',
    }
    application_url = urljoin(base_url, application["self"])
    resp = requests.get(
        url=application_url,
        headers=headers,
        verify=False)
    resp.raise_for_status()
    return resp.json()


def dump_application(base_url, name):
    application = get_application(base_url, name)
    logging.debug("application %s", json.dumps(application, indent=2))
    for inbound_protocol_ref in application['inboundProtocols']:
        headers = {
            'Authorization': f'Basic {b64encode("admin:admin".encode("utf-8")).decode("ascii")}',
            'Content-Type': 'application/json',
        }
        url = urljoin(base_url, inbound_protocol_ref["self"])
        resp = requests.get(
            url=url,
            headers=headers,
            verify=False)
        resp.raise_for_status()
        inbound_protocol = resp.json()
        logging.debug("application %s inbound protocol %s: %s",
            name,
            inbound_protocol_ref['type'],
            json.dumps(inbound_protocol, indent=2))


def create_application(base_url, config):
    """
    create the oidc application and return its details as:

        {
            "id": "",
            "name": "",
            "client_id": "",
            "client_secret": "".
        }
    """

    name = config['name']
    headers = {
        'Authorization': f'Basic {b64encode("admin:admin".encode("utf-8")).decode("ascii")}',
        'Content-Type': 'application/json',
    }
    applications_url = f'{base_url}/api/server/v1/applications'

    # create the application.
    # NB when the application already exists, it will not be updated.
    application = find_application(base_url, name)
    if application:
        id = application['id']
        logging.debug('skipping the application %s creation as it already exists with id %s', name, id)
    else:
        # see https://is.docs.wso2.com/en/latest/apis/application-rest-api/
        # NB this returns an HTTP 201 Created with a location header that has the
        #    application id.
        #    e.g. Location: https://wso2is.test:9443/t/carbon.super/api/server/v1/applications/662c342f-fa69-4f28-ac4b-16bc9a40465d
        logging.debug('creating the application %s', name)
        resp = requests.post(
            url=applications_url,
            headers=headers,
            json=config,
            verify=False)
        resp.raise_for_status()
        location = urlparse(resp.headers['Location'])
        id = location.path.split('/')[-1]
        logging.debug('created the application %s with id %s', name, id)

    # return the application details.
    # see GET ​/applications​/{applicationId}​/inbound-protocols​/oidc
    #     at https://is.docs.wso2.com/en/latest/apis/application-rest-api/
    resp = requests.get(
        url=f'{applications_url}/{id}/inbound-protocols/oidc',
        headers=headers,
        verify=False)
    resp.raise_for_status()
    response = resp.json()
    return {
        "id": id,
        "name": name,
        "client_id": response['clientId'],
        "client_secret": response['clientSecret'],
    }


def update_current_user(base_url, config):
    """
    update the current user.
    """
    headers = {
        'Authorization': f'Basic {b64encode("admin:admin".encode("utf-8")).decode("ascii")}',
        'Content-Type': 'application/json',
    }
    resp = requests.patch(
        url=f'{base_url}/scim2/Me',
        headers=headers,
        json={
            "Operations": [
                {
                    "op": "add",
                    "value": config
                }
            ],
            "schemas": [
                "urn:ietf:params:scim:api:messages:2.0:PatchOp"
            ]
        },
        verify=False)
    resp.raise_for_status()


def dump_current_user(base_url):
    headers = {
        'Authorization': f'Basic {b64encode("admin:admin".encode("utf-8")).decode("ascii")}',
        'Content-Type': 'application/json',
    }
    resp = requests.get(
        url=f'{base_url}/scim2/Me',
        headers=headers,
        verify=False)
    resp.raise_for_status()
    response = resp.json()
    logging.debug("user: %s", json.dumps(response, indent=2))


def init_main(args):
    # wait for wso2is to be available.
    wait_for_ready(args.base_url)

    # update the admin user.
    u = urlparse(args.base_url)
    domain = u.hostname
    update_current_user(
        args.base_url,
        {
            "emails": [
                f"admin@{domain}"
            ],
            "name": {
                "givenName": "Administrator",
                "familyName": "Doe"
            },
            "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": {
                "country": "Portugal",
                "emailVerified": True,
            },
        })

    # create the example-go-confidential application and
    # save its details at ../tmp/example-go-confidential.json.
    application = create_application(
        args.base_url,
        {
            "templateId": "b9c5e11e-fc78-484b-9bec-015d247561b8",
            "name": "example-go-confidential",
            "inboundProtocolConfiguration": {
                "oidc": {
                    "state": "ACTIVE",
                    "grantTypes": [
                        "authorization_code",
                    ],
                    "accessToken": {
                        "type": "Default",
                        "userAccessTokenExpiryInSeconds": 3600,
                        "applicationAccessTokenExpiryInSeconds": 3600,
                        "bindingType": "sso-session",
                        "revokeTokensWhenIDPSessionTerminated": False,
                        "validateTokenBinding": False,
                    },
                    "pkce": {
                        "mandatory": True,
                        "supportPlainTransformAlgorithm": False
                    },
                    "publicClient": False,
                    "validateRequestObjectSignature": False,
                    "callbackURLs": [
                        "http://example-go-confidential.test:8081/auth/oidc/callback"
                    ],
                    "allowedOrigins": [],
                }
            },
            "authenticationSequence": {
                "type": "DEFAULT"
            },
            "advancedConfigurations": {
                "discoverableByEndUsers": False,
                "skipLoginConsent": True
            },
            "claimConfiguration": {
                "dialect": "LOCAL",
                "subject": {
                    "claim": {
                        "uri": "http://wso2.org/claims/username"
                    },
                    "includeTenantDomain": False,
                    "includeUserDomain": False,
                    "useMappedLocalSubject": False
                },
                "role": {
                    "claim": {
                        "uri": "http://wso2.org/claims/roles"
                    },
                    "includeUserDomain": True
                },
                "requestedClaims": [
                    {
                        "claim": {
                            "uri": "http://wso2.org/claims/username"
                        },
                        "mandatory": True
                    },
                    {
                        "claim": {
                            "uri": "http://wso2.org/claims/emailaddress"
                        },
                        "mandatory": True
                    },
                    {
                        "claim": {
                            "uri": "http://wso2.org/claims/identity/emailVerified"
                        },
                        "mandatory": True
                    },
                    {
                        "claim": {
                            "uri": "http://wso2.org/claims/givenname"
                        },
                        "mandatory": True
                    },
                    {
                        "claim": {
                            "uri": "http://wso2.org/claims/lastname"
                        },
                        "mandatory": True
                    },
                ],
            }
        })
    p = Path(__file__).resolve().parent / '..' / 'tmp' / f'{application["name"]}.json'
    with open(p, 'w') as f:
        json.dump(application, f, indent=2)


def try_main(args):
    dump_current_user(args.base_url)
    dump_application(args.base_url, 'example-go-confidential')
