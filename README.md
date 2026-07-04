# Para-NBody (GSIM Engine)

**A 5D Ternary Lattice Engine for Simulating Discrete Spacetime, Dipole Cosmology, and Self-Regulating Quantum Architecture.**

Para-NBody is a highly concurrent, "headless" physics server written in Go. It operates on the **Geometric Substrate Interaction Model (GSIM)**, replacing continuous calculus-based gravitational equations with a discrete, transactional cellular automaton.

Instead of tracking multi-body vectors across empty space, the engine computes localized network strain and dynamic memory allocation across a higher-dimensional network structure.

---

## 🌌 The Core Architecture: The 5D Ternary Fabric

The universe inside this engine is not a continuous vacuum, but a granular, self-balancing data array.

Side A (+W): Matter / Cosmic Filaments (Valleys)
====================== Equator (W = 0 Baseline) ======================
Side B (-W): Antimatter / Cosmic Voids (Hills)

* **The Grid:** A discrete **3D Ternary Lattice Grid** utilizing a base-3 balanced state structure ($-\!1, 0, +1$) to maintain structural equilibrium.
* **The 5D Node Layout:** Each individual cell acts as a local processor tracking 5 spatial and vector variables:
    * $X, Y, Z$: The visible 3D surface mapping coordinates.
    * $W$: Vertical displacement (Mass-Gauge). $+W$ represents regular matter; $-W$ represents antimatter.
    * $V$: Angular torque (The fundamental Spin vector).
* **The Surface Baseline ($W = 0$):** The observable universe is a relaxed, lowest-energy 3D equatorial slice. Particles naturally seek this interface. Mass is physically manifested as the elastic vertical resistance of a coordinate node being pinned off-axis.
* **The Game Tick:** Time is purely transactional, regulated by an atomic double-buffer swap loop. Information and particles propagate at the absolute terminal limit of **one lattice unit per game tick** ($c$).

---

## 📐 Topology: The 5-Field GSIM Network

The underlying vector coordination and particle field layers ($L_0, L_1, L_2$) map directly to the dimensional vertices of the substrate, anchoring the 4D pillar structures around a localized central observer frame:

![GSIM Engine Topology](image_67929a.jpg)

---

## 💫 Emergent Cosmological Rules

By shifting the computation from a continuous top-down formula to a localized neighbor-processing network, three major cosmological anomalies emerge naturally:

### 1. Inverted Gravitational Lensing (The Void Paradox)
Cosmic voids are not empty space. They are highly structured, dense filaments of stable $-W$ antimatter pushing against the grid from the opposite side. Because a $-W$ peak acts as a geometric "hill" rather than a matter "valley," background photons are **deflected outward (diverged)**. Because automated astronomical algorithms are hardcoded to look for converging paths, these dense structures register to external observers as empty voids.

### 2. Dynamic Unit Allocation ("Reverse Spleef" / Dark Energy)
When a localized spatial gradient between adjacent nodes ($\Delta W = |W_{\text{local}} - W_{\text{neighbor}}|$) exceeds the maximum structural yield threshold ($G_{\text{max}}$), the engine dynamically instantiates a new coordinate node between them to preserve causality.
* **Micro Scale:** This strain relief causes the original energy knot to segment, mirroring **weak force particle decay events**.
* **Macro Scale:** The cumulative addition of these new spatial units from trillions of decay events expands the global grid, naturally driving **Cosmic Expansion (Dark Energy)** from the inside out.

### 3. The Bottling Phase (Mass Condensation)
Conversely, when local energy density ($\rho_E$) over-saturates a cluster of nodes, the engine collapses excess empty coordinate units together, "bottling" the raw kinetic energy into a permanent, highly compressed, stable particle knot anchored in its respective subsurface tier.

---

## 🛠 Tech Stack & Performance Architecture

* **Language:** Go (Golang) — utilizing highly concurrent Go-Routines for parallelized local neighborhood calculations.
* **State Management:** Double-buffered spatial chunk arrays for race-condition-free tick computation.
* **Persistence:** SQLite3 transaction-based ledger auditing (0.8ms/step archival rate).
* **Interface:** REST API (JSON-driven universe initialization and runtime control).

---

## 🚀 Quick Start

### 1. Initialize a Universe
Define your start conditions via a JSON POST request. You can seed specific voxels with Mass-Gauge displacement ($+W$ or $-W$) and initial Angular Torque ($V$).

```bash
curl -X POST "http://localhost:42069/api/v1/pnbody/ini" \
-H "Content-Type: application/json" \
-d '{
  "width": 32, "height": 32, "depth": 32,
  "initial_voxels": [
    {"x": 16, "y": 16, "z": 16, "w": 10.0, "v": 1.5},
    {"x": 20, "y": 20, "z": 20, "w": -10.0, "v": -1.5}
  ]
}'

⚖️ License
MIT License.