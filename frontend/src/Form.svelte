<script>
  import { debounce } from "./utils.service.js";
  import { Replace } from "../wailsjs/go/main/App.js";
  import { results, selected_match } from "./store.js";
  import { get_next_match, map_results } from "./results.service.js";
  export let search_term = "turn",
    dir = "/home/stefan/Projects/nvim-float",
    replace_term = "",
    exclude = "",
    include = "";

  async function replace() {
    /**@type {RipgrepResultApi}*/
    await Replace($selected_match, "utils", "foo");
  }

  function foo(event) {
    if (event.key === "Enter") {
      console.log("replacing");
      replace();
    }
  }
</script>

<form on:submit|preventDefault={replace}>
  <div class="flex">
    <div class="w-full">
      <input
        autofocus
        on:keyup={async ({ target: { value } }) =>
          (search_term = await debounce(value))}
        value={search_term}
        type="text"
        placeholder="Search..."
        class="w-full input input-primary bg-surface0 text-text rounded-sm px-2 py-1"
      />
    </div>
    <div class="w-full pl-2">
      <input
        on:keyup={foo}
        bind:value={replace_term}
        type="text"
        placeholder="Replace..."
        class="w-full input input-primary bg-surface0 text-text rounded-sm px-2 py-1"
      />
    </div>
  </div>
  <div class="py-2 flex">
    <div class="w-full">
      <input
        bind:value={dir}
        type="text"
        placeholder="Search..."
        class="w-full input input-primary bg-surface0 text-text rounded-sm px-2 py-1"
      />
    </div>
    <div class="w-full flex">
      <input
        bind:value={exclude}
        type="text"
        placeholder="eg *service.go,src/**/exclude"
        class="w-full ml-2 input input-primary bg-surface0 text-text rounded-sm px-2 py-1"
      />
      <input
        bind:value={include}
        type="text"
        placeholder="eg *service.go,src/**/include"
        class="w-full ml-2 input input-primary bg-surface0 text-text rounded-sm px-2 py-1"
      />
    </div>
  </div>
</form>
