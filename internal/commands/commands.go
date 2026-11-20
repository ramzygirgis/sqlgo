package commands

func Generate() error {
	root, fsys, proj, err := findProject()
	if err != nil {
		return err
	}

	cat, err := loadCatalog(fsys, proj)
	if err != nil {
		return err
	}

	analyzedPackages, err := analyzePackages(fsys, proj, cat)
	if err != nil {
		return err
	}

	outputs, err := render(proj, cat, analyzedPackages)
	if err != nil {
		return err
	}

	stale, err := findStale(fsys, outputs)
	if err != nil {
		return err
	}

	return write(root, outputs, stale)
}
