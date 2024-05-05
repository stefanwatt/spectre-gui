<script>
  import { debounce } from "./utils.service.js";
  export let search_term = "turn",
    dir = "/home/stefan/Projects/nvim-float",
    replace_term = "",
    exclude_filter = "",
    include_filter = "";
</script>

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
      bind:value={exclude_filter}
      type="text"
      placeholder="eg *service.go,src/**/exclude"
      class="w-full ml-2 input input-primary bg-surface0 text-text rounded-sm px-2 py-1"
    />
    <input
      bind:value={include_filter}
      type="text"
      placeholder="eg *service.go,src/**/include"
      class="w-full ml-2 input input-primary bg-surface0 text-text rounded-sm px-2 py-1"
    />
  </div>
</div>
