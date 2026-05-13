# Para-NBody (GSIM Engine)

**A 12-Column Substrate Engine for Simulating Cosmic Chemistry and Lattice-Based Physics.**

Para-NBody is a "Headless" physics server written in Go, designed to explore the **GSIM (Geometric Substrate Interaction Model)**. Unlike traditional N-body simulations that rely on pure Newtonian $1/r^2$ gravity, this engine utilizes a discrete 12-column ledger to simulate the repulsive and attractive "charges" that form the Cosmic Web.

## 🌌 The Theory: Cosmic Chemistry

The core hypothesis of this project is that the Large Scale Structure of the universe (filaments and voids) behaves like a molecular lattice.

* **Matter/Antimatter Dipoles:** Interactions are orientation-dependent.
* **Inverted Orbits:** Galaxies are funneled through filaments by the repulsive pressure of Voids.
* **The GSIM Shunt:** Systematic "rounding errors" in the substrate are shunted into the $V_5$ column, representing the expansion of space.

---

## 🛠 Tech Stack

* **Language:** Go (Golang)
* **Persistence:** SQLite3 (Transaction-based ledger auditing)
* **Interface:** REST API (JSON-driven)
* **Architecture:** Spatial Chunking for scalable grid management

---

## 🚀 Quick Start

### 1. Initialize a Universe

Define your start conditions via a JSON POST request. You can set specific voxels as Matter ($+V_3$) or Antimatter ($-V_3$).

```bash
curl -X POST "http://localhost:42069/api/v1/pnbody/ini" \
-H "Content-Type: application/json" \
-d '{
  "width": 20, "height": 20, "depth": 20,
  "initial_voxels": [
    {"x": 8, "y": 8, "z": 8, "v3": 100},
    {"x": 12, "y": 12, "z": 12, "v3": -100}
  ]
}'

```

### 2. Evolve the Substrate

Run the simulation for a specific number of steps. The engine processes physics and archives the state to SQLite.

```bash
curl -X POST "http://localhost:42069/api/v1/pnbody/{universe_id}/run?count=1000"

```

---

## 📊 Data Auditing

The engine records every step into a SQLite database (`.store`). You can audit the "Pressure" of the universe using standard SQL:

```sql
-- View the highest Expansion Pressure (V5) zones
SELECT x, y, z, v5 FROM chunk_0 ORDER BY v5 DESC LIMIT 10;

```

---

## 🏗 Roadmap

* [x] **Core Substrate:** 12-column multivector implementation.
* [x] **Persistence:** High-speed SQLite archival (0.8ms/step).
* [ ] **Spatial Chunking:** Multi-table distribution for large-scale grids.
* [ ] **Training Architecture:** Interface for AI model feedback loops.
* [ ] **Public Repo:** Open-source physics core for academic peer review.

---

## ⚖️ License

MIT License.

> "The universe isn't a vacuum; it's a circuit."

---