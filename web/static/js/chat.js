function openChatCreateGroupModal() {
	const dialog = document.getElementById("chat-create-group-dialog");
	if (dialog) {
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
	}
}

function syncChatActiveRoom() {
	const url = new URL(window.location.href);
	const room = url.searchParams.get("room");
	const input = document.getElementById("chat-active-room");
	if (input && room) {
		input.value = room;
	}
}

document.addEventListener("DOMContentLoaded", () => {
	syncChatActiveRoom();

	document.body.addEventListener("htmx:afterSwap", (event) => {
		if (event.detail.target && event.detail.target.id === "chat-main") {
			syncChatActiveRoom();
		}
	});

	document.body.addEventListener("htmx:pushedIntoHistory", () => {
		syncChatActiveRoom();
	});
});
