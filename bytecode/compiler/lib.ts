@func yield() Text

@func a() {
  "inside a"
}

@func b() {
  "inside b"
}

@func c() {
  "inside c"
}

@func d() {
  "inside d"
}

@func f() {
  "inside f"
  yield()
}

@func g() {
  "inside g"
  yield()
}