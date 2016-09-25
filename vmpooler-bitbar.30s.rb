#!/usr/bin/env ruby

require 'erb'
require 'commander'
require 'vmfloaty/auth'
require 'vmfloaty/conf'
require 'vmfloaty/pooler'
require 'vmfloaty/utils'
require 'vmfloaty/version'

class VmpoolerBitbar
  include Commander::Methods
  # Minimum supported vmfloaty version

  def run
    program :name, 'vmpooler-bitbar'
    program :version, '1.0.0'
    program :description, 'VMPooler BitBar plugin based on vmfloaty.'

    # Check vmfloaty library version
    min_vmfloaty_version = '0.7.0'
    if Gem::Version.new(Version.get) < Gem::Version.new(min_vmfloaty_version)
      puts '🔥 Update vmfloaty',
            '---',
            "Please update vmfloaty to a version > #{min_vmfloaty_version}",
            "Current version is #{Version.get}",
            '---',
            'Refresh... | refresh=true'
      exit 1
    end



    config = Conf.read_config
    vmpooler_url = config['url']
    token = config['token']

    this_script = File.expand_path $0
    warning_timeleft_threshhold = 1
    extend_lifetime_hours = 2

    def system_notification(message, title="vmpooler bitbar")
      `osascript -e 'display notification "#{message}" with title "#{title}"' &> /dev/null`
    end

    def generate_tag_hash()
      { created_by: 'vmpooler_bitbar' }
    end

    def copy_menu_text_params(menu_text)
      return "bash=/bin/bash param1=-c param2='echo -n #{menu_text} | pbcopy' terminal=false"
    end

    command :menu do |c|
      c.syntax = 'vmpooler-bitbar menu'
      c.description = 'Prints bitbar menu string'
      c.action do |args, options|

        logo_base64 = 'R0lGODlhIAAgAPQAAP+uGv+uG/+vG/+uHP+vHP+vHf+vHv+vH/+wHv+wH/+wIP+xIf+xIv+xI/+xJf+yJP+zJ/6yKf60LAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACH5BAEAABMALAAAAAAgACAAAAWW4CSOZGmeaKqubIsSQSwTrhoAA0EAQFCngUFEZOj9UAGaKGE8mpIjps9ZuhUOiZuBWuV5AUquqDAoCxIFsDhVDK9LaQKjQWdA3pP0F4DAN/YACX6Agm9/e4Vrh1+JYnpffWsHOAqVlQ6SOIoLnAsPRQNvN3sCeDcEBQVWpmGTU2tQS02wAxJEs2KnBnqvuYC9eMHCwyshADs='

        # Menu Item Formatting Parameters - see https://github.com/matryer/bitbar#plugin-api
        header_params = "size=22 font='Arial Bold' templateImage=#{logo_base64}"
        submenu_item_font_size = 'size=12'
        submenu_header_font_size = 'size=14'
        submenu_header_params = "font='Arial Bold' #{submenu_header_font_size}"
        fixed_font_params = "font=Menlo-Regular #{submenu_item_font_size}"
        terminal_action_params = "terminal=true #{submenu_item_font_size}"
        refresh_action_params = "terminal=false refresh=true #{submenu_item_font_size}"
        disabled_action_params = "#{submenu_item_font_size}"

        menu_template = <<-EOS
VMs: <%= vms.length %> | color=<%= expiring_soon ? 'red' : 'green' %>
---
vmpooler | <%= header_params %>
---
<% extend_all_cmd = "" %>
<% if !vms.nil? -%>
<%   vms.each do |vm| -%>
<%     remaining_time_colour = vm[:remaining] <= warning_timeleft_threshhold ? 'red' : 'green' -%>
<%= vm[:hostname] %> (<%= vm[:template] %>) | color=<%= remaining_time_colour %> <%= copy_menu_text_params(vm[:fqdn]) %>
-- Action | <%= submenu_header_params %>
-- SSH to VM | href='ssh://root@<%= vm[:fqdn] %>' <%= terminal_action_params %>
-- Delete VM | bash=<%= this_script %> param1=delete param2=<%= vm[:hostname] %> <%= refresh_action_params %>
-----
-- Extend Lifetime (<%= extend_lifetime_hours %>h) | bash=<%= this_script %> param1=extend param2=<%= vm[:hostname] %> <%= refresh_action_params %>
-----
-- Status | <%= submenu_header_params %>
-- <%= vm[:running] %>/<%= vm[:lifetime] %> hours | <%= fixed_font_params %> color=<%= remaining_time_colour %> <%= copy_menu_text_params(vm[:running].to_s + '/' + vm[:lifetime].to_s + 'hours') %>
-- <%= vm[:template] %> | <%= fixed_font_params %> <%= copy_menu_text_params(vm[:template]) %>
-- <%= vm[:ip] %> | <%= fixed_font_params %> <%= copy_menu_text_params(vm[:ip]) %>
<%     next unless vm.key?(:tags) -%>
-----
-- Tags | <%= submenu_header_params %>
<%     vm[:tags].each do |key, value| -%>
-- <%= key %> = <%= value %> | <%= fixed_font_params %> <%= copy_menu_text_params(value) %>
<%     end -%>
<%   end -%>
<% else -%>
You have no running vms
<% end -%>
---
Bulk Actions
-- Action | <%= submenu_header_params %>
-- Delete | bash=<%= this_script %> param1=delete param2=--all <%= refresh_action_params %>
-- Extend Lifetime (<%= extend_lifetime_hours %>h) | bash=<%= this_script %> param1=extend param2=--all <%= refresh_action_params %>
---
New VM
-- OS Variant | <%= submenu_header_params %>
<% menu_templates.each do |submenu, templates| -%>
-- <%= submenu %>
---- Template | <%= submenu_header_params %>
<%   templates.each do |template| -%>
---- <%= template %> | bash=<%= this_script %> param1=get param2=<%= template %> <%= refresh_action_params %>
<%   end -%>
<% end -%>
Refresh... | refresh=true
EOS

        # Check connectivity and get running vms
        begin
          status = Auth.token_status(false, vmpooler_url, token)
        rescue TokenError => msg
          puts '🔥 Token Error',
               '---',
               "#{msg}",
               'Check your ~/.vmfloaty.yml|href=https://github.com/briancain/vmfloaty#vmfloaty-dotfile',
               'Click for info|href=https://github.com/briancain/vmfloaty#vmfloaty-dotfile',
               '---',
               'Refresh... | refresh=true'
          exit 1
        end

        vms = {}

        if status[token].key?('vms')
          floathosts = status[token]['vms']['running']

          vms = floathosts.map { |x| {:hostname => x } }

          expiring_soon = false

          unless vms.nil?
            # Build hash of vms and details
            vms.each do |vm|
              details = Pooler.query(false, vmpooler_url, vm[:hostname])[vm[:hostname]]
              vm[:template] = details['template']
              vm[:ip] = details['ip']
              vm[:domain] = details['domain']
              vm[:fqdn] = vm[:hostname] + '.' + vm[:domain]
              vm[:lifetime] = details['lifetime']
              vm[:running] = details['running']
              vm[:remaining] = vm[:lifetime] - vm[:running]

              expiring_soon = expiring_soon || vm[:remaining] <= warning_timeleft_threshhold ? true : false

              next unless details.key?('tags')
              vm[:tags] = details['tags']
            end
          end
          # Sort, newer vms first
          vms.sort! { |a, b| a[:running] <=> b[:running] }
        end

        # New VM templates
        vm_templates = Pooler.list(false, vmpooler_url)
        menu_templates = {}
        vm_templates.reject{|template| menu_templates[template.split('-')[0]] = (menu_templates[template.split("-")[0]] ||= []) << template;}

        # Render menu from template
        renderer = ERB.new(menu_template, nil, '-')
        puts(renderer.result(binding))
      end
    end

    command :delete do |c|
      c.syntax = 'vmpooler-bitbar delete [hostname,...]'
      c.description = 'Schedules the deletion of a vm or vms'
      c.option '--all', 'Deletes all vms acquired by a token'
      c.action do |args, options|
        hostnames = args[0]
        delete_all = options.all

        if delete_all
          status = Auth.token_status(false, vmpooler_url, token)
          vms = status[token]['vms']
          unless vms.nil?
            running_vms = vms['running']
            unless running_vms.nil?
              deleted_hosts = []
              errored_hosts = []
              Pooler.delete(false, vmpooler_url, running_vms, token).each do |host,vals|
                if vals['ok'] == true
                  deleted_hosts << host
                else
                  errored_hosts << host
                end
              end

              unless errored_hosts.empty?
                system_notification("Error deleting vm(s): \n#{errored_hosts * ', '}")
              end
              unless deleted_hosts.empty?
                system_notification("Deleting vm(s): \n#{deleted_hosts * ', '}")
              end

            end
          end
          exit 0
        end

        if hostnames.nil?
          abort('You did not provide any hosts to delete')
        else
          hosts = hostnames.split(',')
          response = Pooler.delete(false, vmpooler_url, hosts, token)
          response.each do |host,vals|
            if vals['ok'] == false
              system_notification("Error deleting vm: #{host}")
            else
              system_notification("Deleted vm: #{host}")
            end
          end
        end
        exit 0
      end
    end

    command :extend do |c|
      c.syntax = 'vmpooler-bitbar extend [hostname]'
      c.description = 'Extends the lifetime of a vm by 2 hours'
      c.option '--all', 'Extends the lifetime of all vms acquired by a token by 2 hours'
      c.action do |args, options|
        hostname = args[0]
        extend_all = options.all

        if extend_all
          status = Auth.token_status(false, vmpooler_url, token)
          vms = status[token]['vms']
          unless vms.nil?
            running_vms = vms['running']
            unless running_vms.nil?
              extended_hosts = []
              errored_hosts = []

              running_vms.each do |host|
                current_status = Pooler.query(false, vmpooler_url, host)
                current_lifetime = current_status[host]['lifetime']
                extended_lifetime = current_lifetime + extend_lifetime_hours
                response = Pooler.modify(false, vmpooler_url, host, token, extended_lifetime, {})

                if response['ok'] == true
                  extended_hosts << host
                else
                  errored_hosts << host
                end
              end

              unless errored_hosts.empty?
                system_notification("Error extending vm(s): \n#{errored_hosts * ', '}")
              end
              unless extended_hosts.empty?
                system_notification("Extending vm(s) by 2 hours: \n#{extended_hosts * ', '}")
              end

            end
          end
          exit 0
        end

        if hostname.nil?
          abort('You did not provide a host to extend')
        else
          current_status = Pooler.query(false, vmpooler_url, hostname)
          current_lifetime = current_status[hostname]['lifetime']
          extended_lifetime = current_lifetime + extend_lifetime_hours
          response = Pooler.modify(false, vmpooler_url, hostname, token, extended_lifetime, {})

          if response['ok'] == true
            system_notification("Extended vm: #{hostname}")
          else
            system_notification("Error extending vm: #{hostname}")
          end

        end
        exit 0
      end
    end

    command :get do |c|
      c.syntax = 'vmpooler-bitbar get [template]'
      c.description = 'Gets a vm based on the os template'
      c.action do |args, _options|

        template = args[0]

        if args.nil?
          abort('You did not provide a template to create')
        else
          os_types = Utils.generate_os_hash(args)
          get_response = Pooler.retrieve(false, os_types, token, vmpooler_url)

          if get_response['ok'] == true
            hostname = get_response[template]['hostname']
            system_notification("Created #{template} vm: \n\t#{hostname}")

            tag_hash = generate_tag_hash()
            mod_response = Pooler.modify(false, vmpooler_url, hostname, token, nil, tag_hash)

            if mod_response['ok'] == false
              system_notification("Error tagging #{template} vm: \n\t#{hostname}")
            end

          else
            system_notification('Error creating #{template} vm')
          end

        end
        exit 0
      end
    end

    default_command :menu
    run!

  end
end

VmpoolerBitbar.new.run if $0 == __FILE__