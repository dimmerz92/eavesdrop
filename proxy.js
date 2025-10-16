eventSource = new EventSource("/eavesdrop_sse");

eventSource.onmessage = (event) => {
	if (event.data === "refresh") window.location.reload();
}

eventSource.onerror = (error) => console.error("eavesdrop sse error:", error);
