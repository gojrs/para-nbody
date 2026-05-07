import json
import urllib.request
import time
import random

url = "http://localhost:42069/api/pnbody/"

print("--- GSON Century Batch Starting: 100 Epochs of Discovery ---")

for i in range(1, 101):
    # THE EXPLORATION: Testing the "Max Stress" limits
    repulsion = random.uniform(50.0, 750.0)
    max_speed = random.uniform(5.0, 50.0)

    payload = {
        "n": 2000,
        "steps": 5000,
        "boxSize": 800,
        "unlikeMassRepulsionStrength": repulsion,
        "maxSpeed": max_speed,
        "particleMass": 1.0,
        "label": f"CENTURY_BATCH_{i}"
    }

    req = urllib.request.Request(url, data=json.dumps(payload).encode('utf-8'),
                                 headers={'Content-Type': 'application/json'})

    try:
        with urllib.request.urlopen(req) as response:
            data = json.loads(response.read().decode('utf-8'))
            res = data.get('result', {})
            print(f"[{i}/100] ID: {data.get('id')} | R: {repulsion:0.1f} | Spd: {max_speed:0.1f} | Mass: {res.get('maxMass')} | Survivors: {res.get('finalCount')}")
    except Exception as e:
        print(f"Connection Flux on Run {i}: {e}")

    # 1-second "MacBook Cool-Down" between Big Bangs
    time.sleep(1)

print("--- Multiverse Paving Complete. Check the Ledger tomorrow. ---")
