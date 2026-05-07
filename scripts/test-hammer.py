import requests
import matplotlib.pyplot as plt
import pandas as pd

BASE_URL = "http://localhost:42069"
REPULSION_VALUES = [1.0, 5.0, 10.0, 25.0, 50.0, 100.0, 250.0, 500.0]

def perform_gcc_sweep():
    sweep_data = []
    print(f"🚀 Starting 3D GSON Sweep...")

    for r_val in REPULSION_VALUES:
        payload = {
            "n": 2000,
            "boxSize": 400.0,
            "maxSpeed": 15.0,
            "particleMass": 1.0,
            "steps": 1000,
            "unlikeMassRepulsionStrength": float(r_val),
            "cellSize": 50.0
        }

        try:
            response = requests.post(f"{BASE_URL}/api/pnbody/", json=payload)
            if response.status_code == 200:
                data = response.json()
                # Check if your Go server returns data in "result" or top-level
                res = data.get("result", data)

                sweep_data.append({
                    "repulsion": r_val,
                    "survivors": res.get("finalCount", len(data.get("particles", []))),
                    "max_mass": res.get("maxMass", 0),
                    "id": data.get("id")
                })
                print(f"✅ Strength {r_val}: {sweep_data[-1]['survivors']} survivors | Max Mass: {sweep_data[-1]['max_mass']} (ID: {sweep_data[-1]['id']})")
            else:
                print(f"❌ Strength {r_val} failed: {response.text}")
        except Exception as e:
            print(f"📡 Connection error at {r_val}: {e}")

    return pd.DataFrame(sweep_data)

# 1. Run the Sweep
df = perform_gcc_sweep()

# 2. Plot the Results if we have data
if not df.empty:
    fig, ax1 = plt.subplots(figsize=(10, 6))

    # Left Axis: Survivors
    ax1.set_xlabel('Unlike-Mass Repulsion Strength')
    ax1.set_ylabel('Survivors (Stability)', color='tab:blue')
    ax1.plot(df['repulsion'], df['survivors'], color='tab:blue', marker='o', label='Survivors')
    ax1.tick_params(axis='y', labelcolor='tab:blue')

    # Right Axis: Max Mass
    ax2 = ax1.twinx()
    ax2.set_ylabel('Max Mass (Structure)', color='tab:red')
    ax2.plot(df['repulsion'], df['max_mass'], color='tab:red', linestyle='--', marker='s', label='Max Mass')
    ax2.tick_params(axis='y', labelcolor='tab:red')

    plt.title("GSON Lab: Structural Equilibrium Search (3D)")
    fig.tight_layout()
    plt.grid(True, alpha=0.3)
    plt.show()
