module Tritium
  module Scope
    module NodesetModules
      module SelectionMethods
        #
        # The select function is the primary way to organize your
        # script. It represents searching the Nodeset scope its set
        # inside.
        #
        # When the select() is at the root of the document, it searches
        # the entire tree. However, when its nested, it only searches a
        # subtree of nodes.
        #
        #  select("head") {
        #    select("meta") {
        #      remove()
        #    }
        #  }
        #
        # @param [String] subtree_selector An XPath or CSS3 selector
        #
        # @yield [Scope::Nodeset] A Nodeset scope representing all of the matched elements
        # @return [nil]
        def select(selector, &block)
          child_nodeset = @nodeset.search(selector)
          #child_nodeset.each do |node|
          #  puts "Selector #{selector} matched #{node.path()}"
          #end
          if child_nodeset.size > 0
            child_scope = Scope::Nodeset.new(child_nodeset, @root, self)
            child_scope.instance_eval(&block)
          end
        end
  
        # This is a way to address a Nodeset's attributes
        # for modification purposes. There are several different
        # ways to work with this method.
        #
        # Calling this will apply to *every* node in the Nodeset.
        # Please use another select() if you'd like to only work
        # with nodes having a particular attribute.
        #
        # If you request an attribute that doesn't exist, it gets
        # created automatically. 
        #
        # You can yield to the attribute block and get an Attribute
        # scope, as described below.
        #
        # @example Changing <img> tags to <a> tags
        #  # Search the whole document for img tags
        #  select("img") {
        #     # Renames the tag
        #     name("a") 
        #     # Grab its src attribute (or create it)
        #     attribute("src") {
        #       # Open a [Scope::Text] for the name
        #       name {
        #         # Changes the name of the attribute from 'src' to 'href'
        #         set("href")
        #       }
        #       # Open a [Scope::Text] for the value
        #       value {
        #         # rewrites the value with the result of a url filter
        #         set(filter_url(text))
        #       }
        #     }
        #  }
        #
        # *Note*: If you are trying to delete
        # attributes, you could have a massive inefficiency if 
        # you had a huge Nodeset and then created-then-destroyed
        # attributes on all of them. Always select *only* nodes with the 
        # desired attribute before running this function.
        #
        # 
        #
        # @param [String] Name The name of the attribute
        #
        # @yield [Scope::Attribute] A found or newly created Attribute scope
        # @return [nil]
        def attribute(name, value = nil, &block)
          if value.is_a? Array
            unless value.size == @nodeset.size
              raise StandardError.new("Attribute value is the wrong length: #{value.size}. Should be: #{@nodeset.size}")
            end
          end
          if name.is_a?(String)
            @nodeset.each_with_index do |node, i|
              if value.is_a? String
                node[name] = value.to_s
              elsif value.is_a? Array
                node[name] = value[i].to_s
              end

              attribute_node = node.attributes[name]

              # If we don't have an attribute, make one!
              if attribute_node.nil?
                node[name] = ""
                attribute_node = node.attributes[name]
              end

              if block
                attribute_scope = Scope::Attribute.new(attribute_node, @root, self)
                attribute_scope.instance_eval(&block)
              end
            end
          end
        end

        # This is the main way to get access to the inner HTML inside
        # of each of the nodes that are currently selected. 
        # 
        # Remember, if not used correctly, it can be very slow.
        # Only use in very small trees
        #
        # @example Changing the h1 tag to be a link to the root
        #  select("body") {
        #    select("h1") {
        #      html() {
        #        # Completely overwrites the original contents of ALL matching h1's
        #        set("<a href='/'>Moovweb!</a>")
        #      }
        #    }
        #  }
        #
        # @example Replacing only part of the HTML
        #  html() {
        #    # Will replace "Hampton Catlin" with "Hampton 'Amazing' Catlin"
        #    replace("Hampton Catlin", "Hampton 'Amazing' Catlin")
        #
        #    # if this found "Andrew 'Amazing' Farmer" it would replace it with "(Andrew is definitely not Amazing)"
        #    replace(/Andrew '([^']*)' Farmer/, '(Andrew is definitely not \1)')
        #    
        #  }
        #
        # *Note*: A better way to do the above, is to use the insert_tag method as its more efficient
        #
        # Learn more about text manipulation by referencing the [Scope::Text]
        #
        # @yield [Scope::Text] A text manipulation scope that is raw html
        # @return [nil]
        def html(&block)
          @nodeset.each do |node|
            node.inner_html = open_text_scope_with(node.inner_html, &block)
          end
        end
        alias inner html


        # This is how you select an element to pass it to a function.
        # *Note*: that you can select attributes as I have done in the example.
        #
        # @example Setting some divs to have the onclick of their first anchor child.
        # select("div.row") {
        #   attribute("onclick", fetch("href", "a/@href"))
        #   # Now that I've set the onclick, I'll make it work with 'window.location='
        #   attribute("onclick") {
        #     value {
        #       prepend("window.location = '")
        #       append("'")
        #     }
        #   }
        # }
        #
        # @return [Array] An array of elements represented as strings
        def fetch(selector)
          @nodeset.search(selector).collect do |n|
            n
          end
        end

      end
    end
  end
end
