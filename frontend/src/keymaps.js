/** @type {Modifier[]}*/
let mods = []

window.addEventListener("keydown", (event) => {
  switch (event.key) {
    case "Control":
      mods.push("c")
    case "Shift":
      mods.push("s")
    case "Alt":
      mods.push("a")
  }
});

window.addEventListener("keyup", (event) => {
  switch (event.key) {
    case "Control":
      mods = mods.filter(x => x !== "c")
    case "Shift":
      mods = mods.filter(x => x !== "s")
    case "Alt":
      mods.push("a")
  }
});

/**@param {Modifier} mod*/
function is_mod(mod) {
  return mods.includes(mod)
}
