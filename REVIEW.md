# agent review notes

- Nothing under `internal/templates/instance` may depend on being a
  symlink, and `Init` must skip a template at a symlink's path. Windows
  checkouts (the CI `windows` job) turn symlinks into regular files, which
  embed then ships as templates.
