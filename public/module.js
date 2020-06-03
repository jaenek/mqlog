function newModule(topic) {
	return Module({
		arguments: [
			'-d', vars.delim,
			'-n',
			'-mx', vars.marginx,
			'-my', vars.marginy,
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

newModule(topics.value);

topics.addEventListener("change",
	function() { newModule(topics.value) });

marginx.addEventListener("change",
	function() { vars.marginx = marginx.value.toString(); newModule(topics.value)});

marginy.addEventListener("change",
	function() { vars.marginy = marginy.value.toString(); newModule(topics.value)});
