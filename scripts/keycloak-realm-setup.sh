#!/bin/bash

set -e

KEYCLOAK_URL="http://localhost:8080"
ADMIN_USER="admin"
ADMIN_PASSWORD="admin"
REALM_NAME="direito-lux"

echo "Waiting for Keycloak to be ready..."
until curl -s -f -o /dev/null "${KEYCLOAK_URL}"; do
    echo "Waiting for Keycloak..."
    sleep 5
done
echo "Keycloak is ready!"

echo "Getting admin token..."
ADMIN_TOKEN=$(curl -s -X POST "${KEYCLOAK_URL}/realms/master/protocol/openid-connect/token" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -d "username=${ADMIN_USER}" \
    -d "password=${ADMIN_PASSWORD}" \
    -d "grant_type=password" \
    -d "client_id=admin-cli" | jq -r '.access_token')

if [ -z "$ADMIN_TOKEN" ] || [ "$ADMIN_TOKEN" = "null" ]; then
    echo "Failed to get admin token"
    exit 1
fi

echo "Creating realm: ${REALM_NAME}..."
curl -s -X POST "${KEYCLOAK_URL}/admin/realms" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H "Content-Type: application/json" \
    -d @- <<EOF
{
    "realm": "${REALM_NAME}",
    "enabled": true,
    "displayName": "Direito Lux",
    "sslRequired": "external",
    "registrationAllowed": false,
    "registrationEmailAsUsername": true,
    "rememberMe": true,
    "verifyEmail": true,
    "loginWithEmailAllowed": true,
    "duplicateEmailsAllowed": false,
    "resetPasswordAllowed": true,
    "editUsernameAllowed": false,
    "bruteForceProtected": true,
    "permanentLockout": false,
    "maxFailureWaitSeconds": 900,
    "minimumQuickLoginWaitSeconds": 60,
    "waitIncrementSeconds": 60,
    "quickLoginCheckMilliSeconds": 1000,
    "maxDeltaTimeSeconds": 43200,
    "failureFactor": 30,
    "defaultSignatureAlgorithm": "RS256",
    "offlineSessionMaxLifespanEnabled": false,
    "offlineSessionMaxLifespan": 5184000,
    "clientSessionIdleTimeout": 0,
    "clientSessionMaxLifespan": 0,
    "clientOfflineSessionIdleTimeout": 0,
    "clientOfflineSessionMaxLifespan": 0,
    "accessTokenLifespan": 300,
    "accessTokenLifespanForImplicitFlow": 900,
    "ssoSessionIdleTimeout": 1800,
    "ssoSessionMaxLifespan": 36000,
    "ssoSessionIdleTimeoutRememberMe": 0,
    "ssoSessionMaxLifespanRememberMe": 0,
    "offlineSessionIdleTimeout": 2592000,
    "accessCodeLifespan": 60,
    "accessCodeLifespanUserAction": 300,
    "accessCodeLifespanLogin": 1800,
    "actionTokenGeneratedByAdminLifespan": 43200,
    "actionTokenGeneratedByUserLifespan": 300,
    "oauth2DeviceCodeLifespan": 600,
    "oauth2DevicePollingInterval": 5,
    "internationalizationEnabled": true,
    "supportedLocales": ["pt-BR", "en"],
    "defaultLocale": "pt-BR",
    "browserFlow": "browser",
    "registrationFlow": "registration",
    "directGrantFlow": "direct grant",
    "resetCredentialsFlow": "reset credentials",
    "clientAuthenticationFlow": "clients"
}
EOF

echo "Creating client: direito-lux-app..."
curl -s -X POST "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/clients" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H "Content-Type: application/json" \
    -d @- <<EOF
{
    "clientId": "direito-lux-app",
    "name": "Direito Lux Application",
    "description": "Main application client for Direito Lux",
    "rootUrl": "http://localhost:3000",
    "adminUrl": "http://localhost:3000",
    "baseUrl": "/",
    "surrogateAuthRequired": false,
    "enabled": true,
    "alwaysDisplayInConsole": false,
    "clientAuthenticatorType": "client-secret",
    "redirectUris": [
        "http://localhost:3000/*",
        "http://localhost:8080/*"
    ],
    "webOrigins": [
        "http://localhost:3000",
        "http://localhost:8080"
    ],
    "notBefore": 0,
    "bearerOnly": false,
    "consentRequired": false,
    "standardFlowEnabled": true,
    "implicitFlowEnabled": false,
    "directAccessGrantsEnabled": true,
    "serviceAccountsEnabled": true,
    "publicClient": false,
    "frontchannelLogout": false,
    "protocol": "openid-connect",
    "attributes": {
        "saml.assertion.signature": "false",
        "saml.force.post.binding": "false",
        "saml.multivalued.roles": "false",
        "saml.encrypt": "false",
        "saml.server.signature": "false",
        "saml.server.signature.keyinfo.ext": "false",
        "exclude.session.state.from.auth.response": "false",
        "saml_force_name_id_format": "false",
        "saml.client.signature": "false",
        "tls.client.certificate.bound.access.tokens": "false",
        "saml.authnstatement": "false",
        "display.on.consent.screen": "false",
        "saml.onetimeuse.condition": "false"
    },
    "authenticationFlowBindingOverrides": {},
    "fullScopeAllowed": true,
    "nodeReRegistrationTimeout": -1,
    "defaultClientScopes": [
        "web-origins",
        "role_list",
        "profile",
        "roles",
        "email"
    ],
    "optionalClientScopes": [
        "address",
        "phone",
        "offline_access",
        "microprofile-jwt"
    ]
}
EOF

echo "Creating roles..."
declare -a roles=("admin" "lawyer" "client" "secretary")
for role in "${roles[@]}"; do
    echo "Creating role: ${role}"
    curl -s -X POST "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/roles" \
        -H "Authorization: Bearer ${ADMIN_TOKEN}" \
        -H "Content-Type: application/json" \
        -d "{\"name\": \"${role}\", \"description\": \"${role} role\"}"
done

echo "Creating groups..."
declare -a groups=("Administrators" "Lawyers" "Clients" "Staff")
for group in "${groups[@]}"; do
    echo "Creating group: ${group}"
    curl -s -X POST "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/groups" \
        -H "Authorization: Bearer ${ADMIN_TOKEN}" \
        -H "Content-Type: application/json" \
        -d "{\"name\": \"${group}\"}"
done

echo "Creating admin user..."
curl -s -X POST "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/users" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H "Content-Type: application/json" \
    -d @- <<EOF
{
    "username": "admin@direitolux.com",
    "email": "admin@direitolux.com",
    "emailVerified": true,
    "enabled": true,
    "firstName": "Admin",
    "lastName": "User",
    "realmRoles": ["admin"],
    "credentials": [{
        "type": "password",
        "value": "admin123",
        "temporary": false
    }]
}
EOF

echo "Creating sample lawyer user..."
curl -s -X POST "${KEYCLOAK_URL}/admin/realms/${REALM_NAME}/users" \
    -H "Authorization: Bearer ${ADMIN_TOKEN}" \
    -H "Content-Type: application/json" \
    -d @- <<EOF
{
    "username": "advogado@direitolux.com",
    "email": "advogado@direitolux.com",
    "emailVerified": true,
    "enabled": true,
    "firstName": "João",
    "lastName": "Silva",
    "realmRoles": ["lawyer"],
    "credentials": [{
        "type": "password",
        "value": "lawyer123",
        "temporary": false
    }]
}
EOF

echo "Creating identity providers configuration..."
cat > /tmp/idp-google.json <<EOF
{
    "alias": "google",
    "displayName": "Google",
    "providerId": "google",
    "enabled": false,
    "trustEmail": true,
    "storeToken": false,
    "addReadTokenRoleOnCreate": false,
    "linkOnly": false,
    "firstBrokerLoginFlowAlias": "first broker login",
    "config": {
        "clientId": "\${google.client.id}",
        "clientSecret": "\${google.client.secret}",
        "syncMode": "IMPORT",
        "useJwksUrl": "true"
    }
}
EOF

echo "Creating authentication flow for 2FA..."
cat > /tmp/auth-flow-2fa.json <<EOF
{
    "alias": "browser-with-2fa",
    "description": "Browser flow with 2FA",
    "providerId": "basic-flow",
    "topLevel": true,
    "builtIn": false,
    "authenticationExecutions": []
}
EOF

echo "Realm setup completed successfully!"
echo ""
echo "Realm: ${REALM_NAME}"
echo "Admin user: admin@direitolux.com / admin123"
echo "Lawyer user: advogado@direitolux.com / lawyer123"
echo ""
echo "Access Keycloak Admin Console at: ${KEYCLOAK_URL}/admin"
echo "Access Keycloak Account Console at: ${KEYCLOAK_URL}/realms/${REALM_NAME}/account"