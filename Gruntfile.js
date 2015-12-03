module.exports = function(grunt) {

	// Load grunt tasks
	require('load-grunt-tasks')(grunt);

	// Project configuration
	grunt.initConfig({
		pkg: grunt.file.readJSON('package.json'),
		dirs: {
			output: 'dist',
		},
		concat: {
			'bingo-js': {
				src: ['resources/js/**.js'],
				dest: '<%= dirs.output %>/static/js/bingo.js'
			},
			'vendor-js': {
				src: [
					'resources/vendor/js/jquery.js', // jquery needs to be loaded before bootstrap
					'resources/vendor/js/bootstrap.js',
					'resources/vendor/js/highlight.pack.js',
					'resources/vendor/js/he.js',
					'resources/vendor/js/sjcl.js'
				],
				dest: '<%= dirs.output %>/static/js/vendor.js'
			},
			'vendor-css': {
				src: [
					'resources/vendor/css/bootstrap.css',
					'resources/vendor/css/highlight/default.css'
				],
				dest: '<%= dirs.output %>/static/css/vendor.css'
			}
		},
		uglify: {
			options: {
				banner: '/*! <%= pkg.name %> <%= grunt.template.today("yyyy-mm-dd") %> */\n',
				sourceMap: true,
				compress: true
			},
			bingo: {
				src: '<%= dirs.output %>/static/js/bingo.js',
				dest: '<%= dirs.output %>/static/js/bingo.min.js'
			},
			vendor: {
				src: '<%= dirs.output %>/static/js/vendor.js',
				dest: '<%= dirs.output %>/static/js/vendor.min.js'
			}
		},
		postcss: {
			options: {
				map: {
					inline: true,
				},
				processors: [
					require('cssnext')(), // Embrace the future
					require('stylelint')({
						"extends": "stylelint-config-suitcss",
						"rules": {
							"indentation": [2, "tab", {
								"except": ["value"]
							}]
						}
					}),
					require('cssnano')(), // Minify
				]
			},
			bingo: {
				src: ['resources/css/**.css'],
				dest: '<%= dirs.output %>/static/css/bingo.min.css',
			}
		},
		cssnano: {
			options: {
				sourcemap: true
			},
			vendor: {
				src: '<%= dirs.output %>/static/css/vendor.css',
				dest: '<%= dirs.output %>/static/css/vendor.min.css'
			},
		},
		copy: {
			views: {
				files: [
					{
						expand: true,
						cwd: 'resources/views/',
						src: ['**'],
						dest: '<%= dirs.output %>/views/',
					}
				],
			},
			conf: {
				files: [
					{
						expand: true,
						cwd: 'resources/conf/',
						src: ['**'],
						dest: '<%= dirs.output %>/conf/',
					}
				],
			},
		},
		watch: {
			'bingo-js': {
				files: ['resources/js/**/*.js'],
				tasks: ['bingo-js'],
			},
			'vendor-js': {
				files: ['resources/vendor/js/**/*.js'],
				tasks: ['vendor-js'],
			},
			'bingo-css': {
				files: ['resources/css/**/*.css'],
				tasks: ['bingo-css'],
			},
			'vendor-css': {
				files: ['resources/vendor/**/*.css'],
				tasks: ['vendor-css'],
			},
			views: {
				files: ['resources/views/**/*.html'],
				tasks: ['copy:views'],
			},
			conf: {
				files: ['resources/conf/**/*'],
				tasks: ['copy:conf'],
			}
		},
		clean: {
			all: ['dist'],
			js: ['dist/static/js'],
			css: ['dist/static/css'],
			views: ['dist/views'],
			conf: ['dist/conf']
		}
	});

	// Javascript
	grunt.registerTask('bingo-js', ['concat:bingo-js', 'uglify:bingo']);
	grunt.registerTask('vendor-js', ['concat:vendor-js', 'uglify:vendor']);
	grunt.registerTask('js', ['bingo-js', 'vendor-js']);

	// Styles
	grunt.registerTask('bingo-css', ['postcss']);
	grunt.registerTask('vendor-css', ['concat:vendor-css', 'cssnano']);
	grunt.registerTask('css', ['bingo-css', 'vendor-css']);

	// Default
	grunt.registerTask('default', ['js', 'css', 'copy']);

};
