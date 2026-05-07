import requests
import matplotlib.pyplot as plt
import pandas as pd

# CONFIG: Points to your Para-NBody Master Branch
BASE_URL = "http://localhost:42069"
REPULSION_VALUES = [1.0, 5.0, 10.0, 25.0, 50.0, 100.0, 250.0, 500.0]


def perform_gcc_sweep():
    sweep_data = []
    print(f"🚀 Starting 3D GSON Sweep...")

    for r_val in REPULSION_VALUES:
        ppayload = {
            "n": 500,  # Fewer particles (less chaos)
            "boxSize": 1000.0,  # MUCH larger box (more room to breathe)
            "maxSpeed": 2.0,  # Slow them down so they don't fly out instantly
            "steps": 500,  # Shorter run to see where they are
            "unlikeMassRepulsionStrength": float(r_val),
            "particleMass": 1.0
        }

        try:
            response = requests.post(f"{BASE_URL}/api/pnbody/", json=payload)
            if response.status_code == 200:
                data = response.json()

                # WHACK THE MOLE: Extract from the 'result' object
                res = data.get("result", {})

                survivors = res.get("finalCount", 0)
                max_mass = res.get("maxMass", 0.0)
                run_id = data.get("id")

                sweep_data.append({
                    "repulsion": r_val,
                    "survivors": survivors,
                    "max_mass": max_mass,
                    "id": run_id
                })
                print(f"✅ Strength {r_val}: {survivors} survivors | Max Mass: {max_mass} (ID: {run_id})")
            else:
                print(f"❌ Strength {r_val} failed: {response.text}")
        except Exception as e:
            print(f"📡 Connection error at {r_val}: {e}")

    return pd.DataFrame(sweep_data)


# Execution and Plotting
if __name__ == "__main__":
    df = perform_gcc_sweep()

    if not df.empty:
        fig, ax1 = plt.subplots(figsize=(10, 6))

        # Blue Line: Survivors
        ax1.set_xlabel('Unlike-Mass Repulsion Strength')
        ax1.set_ylabel('Survivors (Stability)', color='tab:blue')
        ax1.plot(df['repulsion'], df['survivors'], color='tab:blue', marker='o', label='Survivors')
        ax1.tick_params(axis='y', labelcolor='tab:blue')

        # Red Line: Structure (Max Mass)
        ax2 = ax1.twinx()
        ax2.set_ylabel('Max Mass (Structure)', color='tab:red')
        ax2.plot(df['repulsion'], df['max_mass'], color='tab:red', linestyle='--', marker='s', label='Max Mass')
        ax2.tick_params(axis='y', labelcolor='tab:red')

        plt.title("GSON Lab: Structural Equilibrium Search (3D)")
        fig.tight_layout()
        plt.grid(True, alpha=0.3)
        plt.show()
    else:
        print("No data collected to plot.")
