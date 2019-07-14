package output

// TODO: Output writer that looks somthing like
//
// source:  /home/.local/etc
// install: /home/.config
//
// [=> base]     -- bash/bashrc
// [=> composed] -- bash/environment
//  -> composing from base and machines/desktop groups
// [=> removed]  -- bash/bad-filemulti
// [=> compiled] -- bash/complex
//  -> ignoring configs in base and common/work due to override
//  -> override file present in common/work-vm
//  -> composing from machines/crunchydev (spliced at common/work-vm:22)
//
// [=> install script] base/bash.install
//  -> triggered by base/bash/bashrc
//  -> triggered by base/bash/environment

// CLI Interfaceo

// dots {config, install, diff, files, help}

// dots install [filter...]
// dots diff    [filter...]
// dots files   [filter...]

// dots config  {profiles, groups, use, override}
