require 'minitest/autorun'
require_relative '../lib/tritium/config'
require_relative '../lib/tritium/serializer'
require 'tempfile'

class ScriptToObjectTest < MiniTest::Unit::TestCase
  include ::Instruction::InstructionType
  
  def test_var_script
    obj = compile_script("var('a')")
    #puts obj.inspect
    assert obj.name.size > 0, :message => "Must have a filename set"
    assert_equal ::Instruction, obj.root.class
    root = obj.root
    assert_equal BLOCK, root.type
    assert_equal nil, root.value
    assert_equal 1, root.children.size
    var_call = root.children.first
    assert_equal FUNCTION_CALL, var_call.type
    assert_equal "var", var_call.value
    assert_equal [], (var_call.children || [])
    
    literal = var_call.arguments.first
    assert_equal TEXT, literal.type
    assert_equal "a", literal.value
  end
  
  def test_import
    script = compile_script("@import hello.ts")
    import = script.root.children.first
    assert_equal import.value, File.absolute_path(import.value)
    assert import.value != nil
  end
  
  def test_regexp_objects
    script = compile_script("/h\\/i/")
    regexp = script.root.children.first
    assert_equal "regexp", regexp.value
    assert_equal FUNCTION_CALL, regexp.type
    assert_equal 0, (regexp.children || []).size
    assert_equal 1, (regexp.arguments || []).size
    regexp_text = regexp.arguments.first
    assert_equal "(?-mix:h\\/i)", regexp_text.value
    assert_equal TEXT, regexp_text.type
  end
  
  def test_instruction_types
    tests = {"/a/" => FUNCTION_CALL,
             "$a"  => FUNCTION_CALL,
             #"%a"  => LOCAL_VAR,
             "'a'" => TEXT,
             '"a"' => TEXT,
             "@import a" => IMPORT }

    tests.each do |script, type| 
       obj = compile_script(script)
       assert_equal type, obj.root.children.first.type
    end
  end
  
  # ============ Helper methods ===============
  
  def bin
    File.absolute_path(File.join(File.dirname(__FILE__), "../bin/ts2to"))
  end
  
  # Creates a tmp folder and actually compiles the ts file into a marshalled
  # stream that comes back
  def compile_script_to_io(script_string, filename = "main")
    script = Tempfile.new([filename, ".ts"])
    output = Tempfile.new("output_#{filename}")
    File.open(script.path, "w") { |f| f.write(script_string) }
    cmd = "#{bin} #{script.path} #{output.path}"
    #puts cmd.inspect
    `#{cmd}`
    open(output.path).read
  end
  
  def compile_script(script_string, filename = "main")
    obj_stream = compile_script_to_io(script_string, filename)
    ::ScriptObject.decode(obj_stream)
  end
end