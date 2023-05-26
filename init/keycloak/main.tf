# see https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/resources/realm
resource "keycloak_realm" "example" {
  realm                    = "example"
  verify_email             = true
  login_with_email_allowed = true
  reset_password_allowed   = true
  smtp_server {
    host = "mailhog"
    port = 1025
    from = "keycloak@keycloak.test"
  }
}

# see https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/resources/group
resource "keycloak_group" "administrators" {
  realm_id = keycloak_realm.example.id
  name     = "administrators"
}

# see https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/resources/group_roles
resource "keycloak_group_roles" "administrators" {
  realm_id = keycloak_group.administrators.realm_id
  group_id = keycloak_group.administrators.id
  role_ids = [
    keycloak_role.wso2is_administrator.id,
  ]
}

# see https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/resources/group_memberships
resource "keycloak_group_memberships" "administrators" {
  realm_id = keycloak_realm.example.id
  group_id = keycloak_group.administrators.id
  members = [
    keycloak_user.alice.username,
  ]
}

# see https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/resources/user
resource "keycloak_user" "alice" {
  realm_id       = keycloak_realm.example.id
  username       = "alice"
  email          = "alice@keycloak.test"
  email_verified = true
  first_name     = "Alice"
  last_name      = "Doe"
  // NB in a real program, omit this initial_password section and force a
  //    password reset.
  initial_password {
    value     = "alice"
    temporary = false
  }
}

# see https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/resources/saml_client
# TODO require client signature.
resource "keycloak_saml_client" "wso2is" {
  realm_id                  = keycloak_realm.example.id
  description               = "WSO2IS"
  client_id                 = "urn:wso2is.test"
  root_url                  = "https://wso2is.test:9443"
  valid_redirect_uris       = ["/commonauth"]
  client_signature_required = false
}

# see https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/resources/saml_user_property_protocol_mapper
# see https://www.keycloak.org/docs-api/21.1.1/javadocs/org/keycloak/models/UserModel.html
resource "keycloak_saml_user_property_protocol_mapper" "wso2is_username" {
  realm_id                   = keycloak_saml_client.wso2is.realm_id
  client_id                  = keycloak_saml_client.wso2is.id
  name                       = "username"
  user_property              = "username"
  saml_attribute_name        = "username"
  saml_attribute_name_format = "Basic"
}

# see https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/resources/saml_user_property_protocol_mapper
# see https://www.keycloak.org/docs-api/21.1.1/javadocs/org/keycloak/models/UserModel.html
resource "keycloak_saml_user_property_protocol_mapper" "wso2is_email" {
  realm_id                   = keycloak_saml_client.wso2is.realm_id
  client_id                  = keycloak_saml_client.wso2is.id
  name                       = "email"
  user_property              = "email"
  saml_attribute_name        = "email"
  saml_attribute_name_format = "Basic"
}

# see https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/resources/saml_user_property_protocol_mapper
# see https://www.keycloak.org/docs-api/21.1.1/javadocs/org/keycloak/models/UserModel.html
resource "keycloak_saml_user_property_protocol_mapper" "wso2is_emailverified" {
  realm_id                   = keycloak_saml_client.wso2is.realm_id
  client_id                  = keycloak_saml_client.wso2is.id
  name                       = "emailverified"
  user_property              = "emailVerified"
  saml_attribute_name        = "emailverified"
  saml_attribute_name_format = "Basic"
}

# see https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/resources/saml_user_property_protocol_mapper
# see https://www.keycloak.org/docs-api/21.1.1/javadocs/org/keycloak/models/UserModel.html
resource "keycloak_saml_user_property_protocol_mapper" "wso2is_firstname" {
  realm_id                   = keycloak_saml_client.wso2is.realm_id
  client_id                  = keycloak_saml_client.wso2is.id
  name                       = "firstname"
  user_property              = "firstName"
  saml_attribute_name        = "firstname"
  saml_attribute_name_format = "Basic"
}

# see https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/resources/saml_user_property_protocol_mapper
# see https://www.keycloak.org/docs-api/21.1.1/javadocs/org/keycloak/models/UserModel.html
resource "keycloak_saml_user_property_protocol_mapper" "wso2is_lastname" {
  realm_id                   = keycloak_saml_client.wso2is.realm_id
  client_id                  = keycloak_saml_client.wso2is.id
  name                       = "lastname"
  user_property              = "lastName"
  saml_attribute_name        = "lastname"
  saml_attribute_name_format = "Basic"
}

# see https://registry.terraform.io/providers/mrparkers/keycloak/latest/docs/resources/role
resource "keycloak_role" "wso2is_administrator" {
  realm_id  = keycloak_saml_client.wso2is.realm_id
  client_id = keycloak_saml_client.wso2is.id
  name      = "administrator"
}
