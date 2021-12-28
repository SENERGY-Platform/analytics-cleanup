export const environment = {
  production: true,
  gateway: 'API_BASE_URL',
  keycloak: {
    url: 'KEYCLOAK_URL'+'/auth',
    realm: 'KEYCLOAK_REALM',
    clientId: 'KEYCLOAK_CLIENT_ID'
  }
};
