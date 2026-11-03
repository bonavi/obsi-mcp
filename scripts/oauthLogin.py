#!/usr/bin/env python3
"""Проходит вход OAuth как клиент и печатает токены. Только для проверок на тестовом сервере."""
import base64, hashlib, json, os, secrets, sys, urllib.error, urllib.parse, urllib.request

base = sys.argv[1]
redirect = "http://localhost:3118/callback"


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, *args):
        return None


def post(path, data, content_type):
    request = urllib.request.Request(base + path, data=data, headers={"Content-Type": content_type})
    return urllib.request.build_opener(NoRedirect).open(request)


client = json.load(post("/oauth/register", json.dumps({"client_name": "e2e", "redirect_uris": [redirect]}).encode(), "application/json"))
verifier = secrets.token_urlsafe(32)
challenge = base64.urlsafe_b64encode(hashlib.sha256(verifier.encode()).digest()).rstrip(b"=").decode()
form = {"client_id": client["client_id"], "redirect_uri": redirect, "response_type": "code", "code_challenge": challenge,
        "code_challenge_method": "S256", "resource": base, "state": "e2e", "password": os.environ["E2E_PASSWORD"]}
try:
    post("/oauth/authorize", urllib.parse.urlencode(form).encode(), "application/x-www-form-urlencoded")
    sys.exit("no redirect after login")
except urllib.error.HTTPError as error:
    location = error.headers.get("Location")
code = urllib.parse.parse_qs(urllib.parse.urlparse(location).query)["code"][0]
token = {"grant_type": "authorization_code", "code": code, "client_id": client["client_id"], "redirect_uri": redirect, "code_verifier": verifier, "resource": base}
print(post("/oauth/token", urllib.parse.urlencode(token).encode(), "application/x-www-form-urlencoded").read().decode())
