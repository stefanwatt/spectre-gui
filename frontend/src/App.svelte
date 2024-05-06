<script>
  import { Search } from "../wailsjs/go/main/App";
  import Form from "./Form.svelte";
  import Results from "./Results.svelte";
  import { highlight_all, map_results } from "./results.service";
  import { selected_match, results } from "./store";
  import { setup_keymaps } from "./keymaps";
  import { onMount } from "svelte";

  let search_term = "utils";
  let replace_term = "foo";
  let dir = "/home/stefan/Projects/bubbletube";
  let include = "**/*.go";
  let exclude = "**/*.sh";
  /**@type {RipgrepMatch} */
  $: {
    try {
      Search(search_term, dir, include, exclude).then(
        /**@param {RipgrepResultApi} res */ (res) => {
          $selected_match = null;
          const mapped = map_results(res);
          results.set(mapped);
          const matches = mapped[0]?.matches;
          if (!matches?.length || !matches[0]) {
            $selected_match = null;
            return;
          }
          const first_match = matches[0];
          console.assert(!!first_match, first_match);
          $selected_match = first_match;
          setTimeout(() => {
            highlight_all();
          });
        },
      );
    } catch (error) {}
  }

  onMount(() => {
    setup_keymaps();
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
