<script>
	import '$lib/assets/prism.css';
	import { EventsOn } from '$lib/wailsjs/runtime/runtime.js';
	import Form from '$lib/Form.svelte';
	import Results from '$lib/Results.svelte';
	import { setup_keymaps } from '$lib/keymaps';
	import Toast from '$lib/notification/Toast.svelte';
	import { show_toast } from '$lib/notification/notification.service.js';
	import { search } from '$lib/results.service';
	import { onMount } from 'svelte';
	import { fade } from 'svelte/transition';
	import {
		toast,
		search_term,
		replace_term,
		dir,
		include,
		exclude,
		search_flags,
		preserve_case
	} from '$lib/store';
	import {} from '$lib/consts';

	/**@type {App.RipgrepMatch} */
	$: {
		const flags = $search_flags.map((f) => f.text);
		search($search_term, $dir, $include, $exclude, flags, $replace_term, $preserve_case);
	}

	onMount(() => {
		setup_keymaps();

		window.addEventListener('keyup', async (e) => {
			if (e.key === 'Enter') {
			}
		});
		EventsOn('files-changed', () => {
			show_toast('info', 'File replaced');
			const flags = $search_flags.map((f) => f.text);
			console.log('searching with flags ', flags);
			search($search_term, $dir, $include, $exclude, flags, $replace_term, $preserve_case);
		});
	});
</script>

<div
	class="min-w-screen flex h-full min-h-screen w-full flex-col overflow-hidden bg-base px-2 py-4"
>
	<Form></Form>
	<div class="h-0 grow snap-y snap-mandatory overflow-x-hidden overflow-y-scroll pt-2">
		<Results></Results>
	</div>
</div>
{#if $toast}
	<div transition:fade={{ delay: 250, duration: 300 }}>
		<Toast text={$toast.text} level={$toast.level}></Toast>
	</div>
{/if}
