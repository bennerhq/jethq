package app

func parseSettingsSection(raw string) (settingsSection, bool) {
	section, err := settingsSectionString(raw)
	return section, err == nil
}
