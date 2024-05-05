<script>
  import { Search } from "../wailsjs/go/main/App";
  import Form from "./Form.svelte";
  import Results from "./Results.svelte";

  /**@type {RipgrepResultApi} */
  let results;
  let search_term = "utils";
  let replace_term = "utilities";
  let dir = "/home/stefan/Projects/nvim-float";
  let include = "**/*.go";
  let exclude = "**/*.sh";
  /**@type {RipgrepMatch} */
  let selected_match = null;
  $: {
    try {
      Search(search_term, dir, include, exclude).then(
        /**@param {RipgrepResultApi} res */ (res) => {
          selected_match = null;
          results = res;
          const entries = Object.entries(res);
          if (!entries?.length) return;
          /**@type {RipgrepMatch[]} */
          const matches = entries[0][1];
          console.assert(matches.length > 0, matches);
          const first_match = matches[0];
          console.assert(!!first_match, first_match);
          selected_match = first_match;
        },
      );
    } catch (error) {}
  }
</script>

<div
  class="bg-base flex flex-col h-full min-h-screen min-w-screen w-full px-2 py-4 overflow-hidden"
>
  <Form bind:search_term bind:replace_term bind:dir bind:include bind:exclude
  ></Form>
  <div
    class="grow h-0 pt-2 overflow-y-scroll overflow-x-hidden snap-y snap-mandatory"
  >
    <Results {selected_match} {search_term} {replace_term} {results}></Results>
  </div>
</div>
