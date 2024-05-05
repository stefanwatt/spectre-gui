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
  $: {
    try {
      Search(search_term, dir, include, exclude).then(
        /**@param {RipgrepResultApi} res */ (res) => {
          results = res;
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
    <Results {search_term} {replace_term} {results}></Results>
  </div>
</div>
