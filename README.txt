git growth explorer

generates commits and measures how .git grows over time.

usage:
  just                  # 100 commits on disk
  just commit 10        # 10 commits
  just commit 1000      # go big
  just commit 100 ram   # use ramdisk (/dev/shm)
  just clean            # nuke playground
  just clean ram        # nuke ramdisk playground

output:
  sha      commit          size        io  delta
  4be33c2       1       29.012K      12ms
  d786ca6       2       30.301K      11ms  -1ms (green)
  a1b2c3d       3       31.590K      14ms  +3ms (red)
  ...

  growth: ~1.3K/commit, est @ 1M: 1.2G
  time: 1.2s total, 1.1s io (92%), 0.1s overhead

how it works:
  1. creates fresh git repo in playground/ (or /dev/shm for ram mode)
  2. copies yaml files into it
  3. randomly tweaks imageName lines and commits
  4. measures .git size and io time after each commit
