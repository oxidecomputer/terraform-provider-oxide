# /// script
# dependencies = [
#   "httpx",
# ]
# ///

import argparse
import uuid

import httpx


def fetch_token(args):
    # Get auth cookie via password login.
    client = httpx.Client()
    client.post(
        f"{args.host}/v1/login/{args.silo}/local",
        json={"username": args.username, "password": args.password},
    ).raise_for_status()

    # Start the device auth flow.
    u = uuid.uuid4()
    device_resp = httpx.post(f"{args.host}/device/auth", data={"client_id": str(u)})
    device_resp.raise_for_status()
    device_details = device_resp.json()

    # Confirm the device via the authenticated session.
    client.post(
        f"{args.host}/device/confirm", json={"user_code": device_details["user_code"]}
    ).raise_for_status()

    # Fetch the token.
    token_resp = httpx.post(
        f"{args.host}/device/token",
        data={
            "grant_type": "urn:ietf:params:oauth:grant-type:device_code",
            "device_code": device_details["device_code"],
            "client_id": str(u),
        },
    )
    token_resp.raise_for_status()
    return token_resp.json()["access_token"]


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--host", default="http://localhost:12220")
    parser.add_argument("--silo", default="test-suite-silo")
    parser.add_argument("--username", default="test-privileged")
    parser.add_argument("--password", default="oxide")
    args = parser.parse_args()

    print(fetch_token(args))
