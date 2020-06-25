function newModule(topic) {
	if (topic == "") {
		return
	}

	return Module({
		arguments: [
			'-d', vars.delim,
			'-mx', vars.marginx,
			'-my', vars.marginy,
			'-l', vars.visible,
			vars.file
		],
		canvas: (() => document.getElementById('canvas'))(),
		locateFile: function(path, prefix) {
		  if (path.endsWith(".data")) return "topics/" + topic;
		  return prefix + path;
		}
	}).then(() => {
		console.log('Loaded!');
	});
}

var vars = {};
const marginx = document.getElementById("mx");
const marginy = document.getElementById("my");
const visible = document.getElementById("vis");

var parts = window.location.href.replace(/[?&]+([^=&]+)=([^&]*)/gi,
	function(m,key,value) {vars[key] = value;});

if (!vars.file) {
	vars.file='data';
}

if (!vars.delim) {
	vars.delim=';';
}

if (!vars.marginx) {
	vars.marginx = '0';
} else {
	marginx.value = parseInt(vars.marginx);
}

if (!vars.marginy) {
	vars.marginy = '100';
} else {
	marginy.value = parseInt(vars.marginy);
}

if (!vars.visible) {
	vars.visible = '50';
} else {
	visible.value = parseInt(vars.visible);
}

newModule(topics.value);

topics.addEventListener("change",
	function() { newModule(topics.value) });

marginx.addEventListener("change",
	function() { vars.marginx = marginx.value.toString(); newModule(topics.value)});

marginy.addEventListener("change",
	function() { vars.marginy = marginy.value.toString(); newModule(topics.value)});

visible.addEventListener("change",
	function() { vars.visible = visible.value.toString(); newModule(topics.value)});
