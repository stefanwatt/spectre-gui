<script>
  import { EventsOn } from "../wailsjs/runtime/runtime.js";
  import Form from "./Form.svelte";
  import Results from "./Results.svelte";
  import { setup_keymaps } from "./keymaps";
  import Toast from "./notification/Toast.svelte";
  import { show_toast } from "./notification/notification.service.js";
  import { search } from "./results.service";
  import { onMount } from "svelte";
  import { fade } from "svelte/transition";
  import { toast, search_term, dir, include, exclude } from "./store";

  /**@type {RipgrepMatch} */
  $: {
    search($search_term, $dir, $include, $exclude);
  }

  onMount(() => {
    setup_keymaps();

    window.addEventListener("keyup", async (e) => {
      if (e.key === "Enter") {
      }
    });
    EventsOn("files-changed", () => {
      show_toast("info", "File replaced");
      search($search_term, $dir, $include, $exclude);
    });
  });
</script>

<div
  class="bg-base flex flex-col h-full min-h-screen min-w-screen w-full px-2 py-4 overflow-hidden"
>
  <Form></Form>
  <div
    class="grow h-0 pt-2 overflow-y-scroll overflow-x-hidden snap-y snap-mandatory"
  >
    <Results></Results>
  </div>
</div>
{#if $toast}
  <div transition:fade={{ delay: 250, duration: 300 }}>
    <Toast text={$toast.text} level={$toast.level}></Toast>
  </div>
{/if}
