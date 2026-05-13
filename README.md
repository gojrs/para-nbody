# 📂 GCC Project: Simulation Mechanics & Logic Ledger

## 1. The Core Philosophy: "The Floor is Bumpy"

We do not use "Action-at-a-Distance." Particles do not "pull" on each other. Instead:

* **The Membrane:** Think of the simulation space as a 3D floor (the Brane) in a 4D room (the Bulk).
* **Mass Displacement:** Mass acts as a weight. It pushes the floor down (**Valley**) or pulls it up (**Hill**).
* **Motion:** Particles simply "slide" down slopes. They only care about the **Power Level (Slope)** of the "Dot" they are currently standing on.

## 2. The Accounting Architecture (Triple-Contract)

To ensure the simulation is verifiable by both Humans and AI (Smarter Beings), we use three "Truth Sources":

1. **Genesis (YAML):** Defines the initial "Sane Defaults" (Row 1).
2. **Interface (Protobuf):** The binary "Contract" used for high-speed Ref/Worker/LG communication.
3. **History (SQL):** The "Memory" of the universe, storing every "Survivor Pic" and Seed variation.

## 3. The Execution Loop (Ref/Worker/LG)

1. **The Worker (Local Logic):**
* Checks the local "Dot" for slope metadata.
* Moves based on local geometry.
* *Note:* Terminal Zeros = "Level Ground" (Low-E). No slope = no change.


2. **The LifeGuard (The Auditor):**
* **The Death List:** Identifies "Illegal" placements (collisions/tears).
* **Liquidation:** Converts dead particles into **Energy Escrow**.


3. **The Referee (The Global Tick):**
* **The Step:** Syncs the LG Audit with the Worker intents.
* **Re-Investment:** If **Escrow E** exceeds the threshold, the Ref "paves" new dots (Spacetime Expansion).



## 4. Time as a Service Fee (Latency)

Time in this simulation is the latency of the refresh rate.

* Deep Valleys (High Mass) require more "Paving Resolution."
* **Time Dilation:** The extra CPU cycles the Ref needs to process a deep displacement. Time slows down where the map is "heavier."

## 5. Known "Skill Issues" & Debugging

* **Terminal Zeros:** Feature, not a bug. In Low-E states, the Delta ($\Delta$) is 0 because the map is already stable.
* **The Escrow Buffer:** If the simulation stops expanding, the "Escrow" is likely below the **Paving Threshold**. It needs a "Death" to trigger new growth.
* **Runaway Pairs:** Emergent behavior. When a "Hill" meets a "Valley," the resulting slope creates a slingshot effect.
