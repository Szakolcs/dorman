(function () {
	var STORAGE_KEY = "dm-theme-preference";

	function systemTheme() {
		return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
	}

	function applyTheme(mode) {
		var root = document.documentElement;
		if (mode === "system") {
			root.removeAttribute("data-theme");
			return;
		}
		root.setAttribute("data-theme", mode);
	}

	function init() {
		var stored = localStorage.getItem(STORAGE_KEY);
		if (stored === "light" || stored === "dark" || stored === "system") {
			applyTheme(stored);
		} else {
			applyTheme("system");
		}
	}

	function cycleTheme() {
		var stored = localStorage.getItem(STORAGE_KEY) || "system";
		var next = stored === "system" ? "light" : stored === "light" ? "dark" : "system";
		localStorage.setItem(STORAGE_KEY, next);
		applyTheme(next);
		updateToggleLabels();
	}

	function updateToggleLabels() {
		var stored = localStorage.getItem(STORAGE_KEY) || "system";
		var label =
			stored === "system"
				? "Theme: System (" + systemTheme() + ")"
				: "Theme: " + stored.charAt(0).toUpperCase() + stored.slice(1);
		document.querySelectorAll("[data-theme-toggle-label]").forEach(function (el) {
			el.textContent = label;
		});
	}

	window.dmCycleTheme = cycleTheme;

	init();
	document.addEventListener("DOMContentLoaded", updateToggleLabels);
	window
		.matchMedia("(prefers-color-scheme: dark)")
		.addEventListener("change", function () {
			if ((localStorage.getItem(STORAGE_KEY) || "system") === "system") {
				updateToggleLabels();
			}
		});
})();
