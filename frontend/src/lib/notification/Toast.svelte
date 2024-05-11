<script>
	import Error from '$lib/icons/Error.svelte';
	import Info from '$lib/icons/Info.svelte';
	import Success from '$lib/icons/Success.svelte';
	import Warning from '$lib/icons/Warning.svelte';

	/**@type { string }*/
	export let text;
	/**@type { App.NotificationLevel }*/
	export let level = 'info';
</script>

<div class="toast">
	<div
		class:!bg-blue={level === 'info'}
		class:!bg-green={level === 'success'}
		class:!bg-yellow={level === 'warning'}
		class:!bg-red={level === 'error'}
		class="alert !text-base"
	>
		<span>
			{#if level === 'info'}
				<Info></Info>
			{/if}
			{#if level === 'success'}
				<Success></Success>
			{/if}
			{#if level === 'warning'}
				<Warning></Warning>
			{/if}
			{#if level === 'error'}
				<Error></Error>
			{/if}
		</span>
		<span>{text}</span>
	</div>
</div>
