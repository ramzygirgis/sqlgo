package cli

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

	stale, err := findStale(fsys, analyzedPackages) // generated files no longer produced
	if err != nil {
		return err
	}

	// generate models.go, using proj.ModelsInfo
	// generate db.go, just in same dir as models
	return write(root, analyzedPackages, stale)
}
