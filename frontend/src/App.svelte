<script>
  import { EventsOn } from "../wailsjs/runtime/runtime.js";
  import Form from "./Form.svelte";
  import Results from "./Results.svelte";
  import { setup_keymaps } from "./keymaps";
  import { search } from "./results.service";
  import { onMount } from "svelte";

  let search_term = "utils";
  let replace_term = "foo";
  let dir = "/home/stefan/Projects/bubbletube";
  let include = "**/*.go";
  let exclude = "**/*.sh";
  /**@type {RipgrepMatch} */
  $: {
    search(search_term, dir, include, exclude);
  }

  onMount(() => {
    setup_keymaps();
    EventsOn("files-changed", () => {
      search(search_term, dir, include, exclude);
    });
  });
</script>

<div
  class="bg-base flex flex-col h-full min-h-screen min-w-screen w-full px-2 py-4 overflow-hidden"
>
  <Form bind:search_term bind:replace_term bind:dir bind:include bind:exclude
  ></Form>
  <div
    class="grow h-0 pt-2 overflow-y-scroll overflow-x-hidden snap-y snap-mandatory"
  >
    <Results {search_term} {replace_term}></Results>
  </div>
</div>
