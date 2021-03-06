gorest
======

A fork of http://code.google.com/p/gorest/

Documentation: http://godoc.org/github.com/fromkeith/gorest

This Fork
=========

This fork changes the following:

- Response marshallers return the type io.Reader instead of []byte. (Breaking change)
- RegisterRecoveryHandler(handler) added. This allows runtime errors to be handled by the user of the library.
- OverrideLogger(logger) added. Allows the logger used in gorest to be overridden
- RegisterHealthHandler(handler) added. Allows health info (status codes) to be more easily reported.
- Individual endpoints can now specify their own 'produces' tag. So you can have 1 service output many data types.
- Allow Put, Post and Delete requests to return data in the body.
- Applied some enforcement on the 'output' tag, to ensure it is specified. (Unknown if this will break people)
- Generates documentation on your services. See DocumentServices for more info.
- Allows you to redefine default result codes. So if you want 200 for POST you can now set that.

License
=======
Original code is from: http://code.google.com/p/gorest/  Please see their license and headers in files.

Modified code is licensed as:

Copyright (c) 2014, fromkeith
All rights reserved.

Redistribution and use in source and binary forms, with or without modification,
are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this
  list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice, this
  list of conditions and the following disclaimer in the documentation and/or
  other materials provided with the distribution.

* Neither the name of the fromkeith nor the names of its
  contributors may be used to endorse or promote products derived from
  this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.



