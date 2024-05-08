<script>
	/**@type { string }*/
	export let text;
	/**@type { App.NotificationLevel }*/
	export let level;
</script>

<div class="toast">
	<div class="alert alert-{level}">
		<span>{text}</span>
	</div>
</div>
