import json
import os
import sys
import urllib.error
import urllib.request

BASE_URL = os.getenv("BASALTPASS_BASE_URL") or os.getenv("BASALT_BASE_URL", "http://localhost:8101")
API_BASE = f"{BASE_URL}/api/v1/manual"
API_KEY = (os.getenv("BASALTPASS_API_KEY") or os.getenv("BASALT_API_KEY", "")).strip()

FRONTEND_URL = os.getenv("POSTCUBE_FRONTEND_URL", "http://localhost:5116")
BACKEND_URL = os.getenv("POSTCUBE_BACKEND_URL", "http://localhost:8113")
CALLBACK_URL = f"{BACKEND_URL}/api/auth/callback"

if not API_KEY:
    print("ERROR: BASALTPASS_API_KEY is required.")
    sys.exit(1)

headers = {
    "X-Api-Key": API_KEY,
    "Content-Type": "application/json",
}


def request_json(method: str, url: str, payload: dict | None = None) -> dict:
    data = None
    if payload is not None:
        data = json.dumps(payload).encode("utf-8")

    req = urllib.request.Request(url, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req) as response:
            return json.loads(response.read().decode("utf-8"))
    except urllib.error.HTTPError as err:
        body = err.read().decode("utf-8", errors="replace")
        print(f"HTTP {err.code} for {url}: {body}")
        raise


print("Creating Postcube app...")
app_payload = {
    "name": "Postcube",
    "description": "Anonymous question box app",
    "homepage_url": FRONTEND_URL,
    "redirect_uris": [CALLBACK_URL],
    "allowed_origins": [FRONTEND_URL],
}
app_resp = request_json("POST", f"{API_BASE}/apps", app_payload)
print("Create app response:")
print(json.dumps(app_resp, indent=2))

app_id = app_resp.get("data", {}).get("id")
if not app_id:
    print("ERROR: app_id not returned, stop.")
    sys.exit(1)

existing_oauth_clients = app_resp.get("data", {}).get("oauth_clients") or []
if existing_oauth_clients:
    first = existing_oauth_clients[0]
    client_id = first.get("client_id")
    client_secret = first.get("client_secret")
    print("App creation already returned an OAuth client.")
    print("\n" + "=" * 60)
    print("Add these values to Postcube/backend/.env")
    print(f"BASALTPASS_CLIENT_ID={client_id}")
    print(f"BASALTPASS_CLIENT_SECRET={client_secret}")
    print("=" * 60)
    sys.exit(0)

print("Creating Postcube OAuth client...")
oauth_payload = {
    "app_id": app_id,
    "name": "Postcube OAuth",
    "description": "OAuth client for Postcube",
    "redirect_uris": [CALLBACK_URL],
    "allowed_origins": [FRONTEND_URL],
}
oauth_resp = request_json("POST", f"{API_BASE}/oauth/clients", oauth_payload)
print("Create OAuth response:")
print(json.dumps(oauth_resp, indent=2))

client_id = oauth_resp.get("data", {}).get("client_id")
client_secret = oauth_resp.get("data", {}).get("client_secret")

print("\n" + "=" * 60)
print("Add these values to Postcube/backend/.env")
print(f"BASALTPASS_CLIENT_ID={client_id}")
print(f"BASALTPASS_CLIENT_SECRET={client_secret}")
print("=" * 60)
