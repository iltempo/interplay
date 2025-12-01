# Musical Enhancement Ideas

Ideas for making patterns more musical and expressive.

## Implemented ✅
- ✅ Humanization (velocity, timing, gate randomization)
- ✅ Per-step velocity/gate/duration control
- ✅ Variable pattern lengths
- ✅ Swing/Groove timing

## High Priority

### 2. Note Probability ⭐⭐
Each step has a % chance to play - adds variation between loops
- `probability <step> <percent>` - e.g., 70% means plays 7 out of 10 loops
- Creates evolving, less predictable patterns
- Great for adding ghost notes or fills
- **Impact**: High - patterns evolve organically
- **Complexity**: Low - random check before note trigger

### 3. Scale Quantization ⭐
Automatically snap notes to a musical scale
- `scale <root> <type>` - e.g., `scale C minor`, `scale D dorian`
- Ensures everything sounds "in key"
- AI could suggest scales based on pattern
- **Impact**: High - guarantees musical coherence
- **Complexity**: Medium - scale definitions + note mapping

## Medium Priority

### 4. Ratcheting
Repeat a note multiple times within its step (machine-gun effect)
- `ratchet <step> <count>` - play note 2-4 times rapidly
- Common in electronic music for energy/fills
- **Impact**: Medium - adds excitement to specific steps
- **Complexity**: Medium - subdivide step timing

### 5. Conditional Triggers
Step plays only on certain loop iterations
- Every Nth loop, first/last loop only, odd/even loops
- Creates evolving patterns over time
- **Impact**: Medium - long-form pattern evolution
- **Complexity**: Medium - loop counter tracking

### 6. Note Ranges/Pools
Instead of fixed note, randomly pick from a set
- `set 1 pool C3,E3,G3` - randomly plays one of those notes
- Controlled randomness within harmonic boundaries
- **Impact**: Medium - controlled variation
- **Complexity**: Low - random selection from array

### 7. Accent Patterns
Automatically boost velocity on certain steps
- Classic drum machine style (accent every 4th, etc.)
- Creates rhythmic emphasis
- **Impact**: Low - subtle enhancement
- **Complexity**: Low - velocity patterns

## Notes
- Many of these can be combined (e.g., swing + probability + humanization)
- AI mode can help users discover and apply these creatively
- Focus on features that add musicality without complexity
