function openChatCreateGroupModal() {
	const dialog = document.getElementById("chat-create-group-dialog");
	if (dialog) {
		resetChatGroupMemberPicker();
		dialog.showModal();
	}
}

function closeChatCreateGroupModal() {
	const dialog = document.getElementById("chat-create-group-dialog");
	if (dialog) {
		dialog.close();
		const form = dialog.querySelector("form");
		if (form) {
			form.reset();
		}
		resetChatGroupMemberPicker();
	}
}

function resetChatGroupMemberPicker() {
	const available = document.getElementById("chat-group-available");
	const selected = document.getElementById("chat-group-selected");
	const inputs = document.getElementById("chat-group-members-inputs");
	if (!available || !selected || !inputs) {
		return;
	}
	while (selected.firstChild) {
		available.appendChild(selected.firstChild);
	}
	inputs.replaceChildren();
	sortChatMemberList(available);
}

function sortChatMemberList(list) {
	const items = Array.from(list.querySelectorAll(".dm-chat-member-chip"));
	items.sort((a, b) =>
		a.dataset.tenantName.localeCompare(b.dataset.tenantName, undefined, { sensitivity: "base" }),
	);
	items.forEach((item) => list.appendChild(item));
}

function addChatGroupMember(chip) {
	const selected = document.getElementById("chat-group-selected");
	const inputs = document.getElementById("chat-group-members-inputs");
	if (!selected || !inputs || !chip) {
		return;
	}
	selected.appendChild(chip);
	const input = document.createElement("input");
	input.type = "hidden";
	input.name = "members";
	input.value = chip.dataset.tenantId;
	input.dataset.tenantId = chip.dataset.tenantId;
	inputs.appendChild(input);
	sortChatMemberList(selected);
}

function removeChatGroupMember(chip) {
	const available = document.getElementById("chat-group-available");
	const inputs = document.getElementById("chat-group-members-inputs");
	if (!available || !inputs || !chip) {
		return;
	}
	available.appendChild(chip);
	const input = inputs.querySelector(`input[data-tenant-id="${chip.dataset.tenantId}"]`);
	if (input) {
		input.remove();
	}
	sortChatMemberList(available);
}

function syncChatActiveRoom() {
	const url = new URL(window.location.href);
	const room = url.searchParams.get("room");
	const input = document.getElementById("chat-active-room");
	if (input) {
		input.value = room || "";
	}
}

function initChatPage() {
	syncChatActiveRoom();

	document.body.addEventListener("click", (event) => {
		const chip = event.target.closest(".dm-chat-member-chip");
		if (!chip) {
			return;
		}
		if (chip.closest("#chat-group-available")) {
			addChatGroupMember(chip);
		} else if (chip.closest("#chat-group-selected")) {
			removeChatGroupMember(chip);
		}
	});

	document.body.addEventListener("htmx:afterSwap", (event) => {
		if (event.detail.target && event.detail.target.id === "chat-main") {
			syncChatActiveRoom();
		}
	});

	document.body.addEventListener("htmx:pushedIntoHistory", () => {
		syncChatActiveRoom();
	});
}

if (document.readyState === "loading") {
	document.addEventListener("DOMContentLoaded", initChatPage);
} else {
	initChatPage();
}
