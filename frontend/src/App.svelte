<script>
  import { Search } from "../wailsjs/go/main/App";
  import Form from "./Form.svelte";
  import Results from "./Results.svelte";

  /**@type {RipgrepResult} */
  let results;
  let search_term = "utils";
  let dir = "/home/stefan/Projects/nvim-float";
  $: {
    try {
      Search(search_term, dir).then(
        /**@param {RipgrepResult} res */ (res) => {
          results = res;
        },
      );
    } catch (error) {}
  }
</script>

<div
  class="flex flex-col h-full min-h-screen min-w-screen w-full px-2 py-4 overflow-hidden"
>
  <Form bind:search_term bind:dir></Form>
  <div
    class="grow h-0 pt-2 overflow-y-scroll overflow-x-hidden snap-y snap-mandatory"
  >
    <Results {results}></Results>
  </div>
</div>
