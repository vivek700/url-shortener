import requests
import time
import json

BASE_URL = "http://localhost:8080"

def test_url_shortener():
    # Store short codes for later use
    short_codes = []
    
    def print_response(response):
        print("\nStatus Code:", response.status_code)
        print("Response:")
        try:
            print(json.dumps(response.json(), indent=2))
        except:
            print(response.text)
        print("-" * 50)

    # Test 1: Shorten multiple URLs
    test_urls = [
        "https://www.google.com",
        "https://github.com",
        "https://www.youtube.com",
        "https://www.reddit.com",
        "https://www.wikipedia.org"
    ]

    print("\nTest 1: Shortening URLs")
    for url in test_urls:
        response = requests.post(
            f"{BASE_URL}/api/shorten",
            json={"url": url}
        )
        print(f"\nShortening {url}")
        print_response(response)
        
        if response.status_code == 201:
            short_codes.append(response.json()["short_code"])

    # Test 2: Invalid URL
    print("\nTest 2: Testing invalid URL")
    response = requests.post(
        f"{BASE_URL}/api/shorten",
        json={"url": "not-a-valid-url"}
    )
    print_response(response)

    # Test 3: Missing URL in request
    print("\nTest 3: Testing missing URL")
    response = requests.post(
        f"{BASE_URL}/api/shorten",
        json={}
    )
    print_response(response)

    # Test 4: List all URLs
    print("\nTest 4: Listing all URLs")
    response = requests.get(f"{BASE_URL}/api/urls")
    print_response(response)

    # Test 5: Test redirects
    print("\nTest 5: Testing redirects")
    for code in short_codes:
        print(f"\nTesting redirect for code: {code}")
        response = requests.get(
            f"{BASE_URL}/r/{code}",
            allow_redirects=False
        )
        print("Redirect Status:", response.status_code)
        print("Redirect Location:", response.headers.get('Location'))
        print("-" * 50)

    # Test 6: Invalid short code
    print("\nTest 6: Testing invalid short code")
    response = requests.get(f"{BASE_URL}/r/invalid123")
    print_response(response)

    # Test 7: Rate limiting
    print("\nTest 7: Testing rate limiting")
    for i in range(25):  # Try 25 requests (rate limit is 20)
        response = requests.get(f"{BASE_URL}/api/urls")
        print(f"Request {i+1}: Status Code {response.status_code}")
        if response.status_code != 200:
            print_response(response)
            break

if __name__ == "__main__":
    test_url_shortener()