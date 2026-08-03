#!/bin/bash
# Generates the throwaway RSA keypair the fake OIDC provider signs id_tokens with.
# Writes the private key as PEM and its public half as a JWK, so the provider can
# sign and publish a matching JWKS. Both outputs are gitignored - run this once
# after cloning, then `duso run_oauth_tests.du`.
#
# Needs openssl, which ships with macOS and Linux. Not a duso script because duso
# has no exec builtin, deliberately: a runtime whose sandbox story is -no-files
# should not hand scripts a shell.
set -eu
cd "$(dirname "$0")"

KEY=oidc_test_key.pem
JWK=oidc_test_key.jwk.json
KID=test-key-1

if ! command -v openssl >/dev/null 2>&1; then
  echo "error: openssl not found" >&2
  exit 1
fi

openssl genrsa -out "$KEY" 2048 2>/dev/null
chmod 600 "$KEY"

# base64url, per RFC 7517: URL-safe alphabet, padding stripped
b64url() {
  base64 | tr -d '\n' | tr '+/' '-_' | tr -d '='
}

# The modulus comes back as hex; the JWK wants it as base64url of the raw bytes.
# perl rather than xxd, which is not installed everywhere.
MODULUS=$(openssl rsa -in "$KEY" -noout -modulus | sed 's/^Modulus=//')
N=$(printf '%s' "$MODULUS" | perl -pe 's/(..)/chr(hex($1))/ge' | b64url)

# openssl uses 65537 unless told otherwise. Check rather than assume, since a
# different exponent would silently produce a JWK that does not match the key.
EXPONENT=$(openssl rsa -in "$KEY" -noout -text | sed -n 's/.*publicExponent: \([0-9]*\).*/\1/p')
if [ "$EXPONENT" != "65537" ]; then
  echo "error: unexpected public exponent $EXPONENT (expected 65537)" >&2
  exit 1
fi
E="AQAB"

cat > "$JWK" <<EOF
{
  "kty": "RSA",
  "alg": "RS256",
  "use": "sig",
  "kid": "$KID",
  "n": "$N",
  "e": "$E"
}
EOF

echo "wrote $KEY and $JWK"
