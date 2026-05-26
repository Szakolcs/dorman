(function () {
	var attributeOrder = ["sex", "nationality", "faculty", "degree", "age"];

	var attributeLabels = {
		sex: "Sex",
		nationality: "Nationality",
		faculty: "Faculty",
		degree: "Degree",
		age: "Age",
	};

	var strictGroups = attributeOrder.slice();
	var preferences = [];

	function renderContainer(containerId, items) {
		var container = document.getElementById(containerId);
		if (!container) {
			return;
		}
		container.innerHTML = "";
		items.forEach(function (id) {
			var chip = document.createElement("button");
			chip.type = "button";
			chip.className = "dm-mass-assignment-chip";
			chip.dataset.attribute = id;
			chip.textContent = attributeLabels[id] || id;
			chip.addEventListener("click", function () {
				toggleAttribute(id, containerId);
			});
			container.appendChild(chip);
		});
	}

	function syncHiddenInputs() {
		var strictInput = document.getElementById("strict-groups-input");
		var preferencesInput = document.getElementById("preferences-input");
		if (strictInput) {
			strictInput.value = strictGroups.join(",");
		}
		if (preferencesInput) {
			preferencesInput.value = preferences.join(",");
		}
	}

	function renderAll() {
		renderContainer("strict-groups", strictGroups);
		renderContainer("preferences", preferences);
		syncHiddenInputs();
	}

	function toggleAttribute(id, containerId) {
		if (containerId === "strict-groups") {
			strictGroups = strictGroups.filter(function (item) {
				return item !== id;
			});
			preferences.push(id);
		} else {
			preferences = preferences.filter(function (item) {
				return item !== id;
			});
			strictGroups.push(id);
		}
		renderAll();
	}

	document.addEventListener("DOMContentLoaded", renderAll);

	document.body.addEventListener("massAssignmentAlert", function (evt) {
		var message = evt.detail;
		if (message && typeof message === "object" && "value" in message) {
			message = message.value;
		}
		alert(message);
	});
})();
