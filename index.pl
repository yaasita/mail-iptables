#!/usr/bin/perl
use Mojolicious::Lite;
use Mojo::JSON qw(decode_json encode_json);
use Time::Piece;

get '/' => sub {
    my $c = shift;
    my $ip = $c->tx->remote_address;
    {
        my $json = {};
        my $jsonfile = "tmp/ip.json";
        if ( -e $jsonfile ){
            open (my $in,"<", "$jsonfile") or die $!;
            local $/ = undef;
            my $data = <$in>;
            $json = decode_json($data);
        }
        my $t = localtime;
        $json->{$ip} = $t->strftime("%Y%m%d_%H:%M");
        my $i;
        for(sort {$json->{$b} cmp $json->{$a}} keys %{$json}){
            $i++;
            if($i>30){
                delete($json->{$_});
            }
        }
        open (my $wr, ">", $jsonfile) or die $!;
        print $wr encode_json($json);
    }
    system("sudo -Hu root /path/to/create-accept-ip") and die $!;
    $c->render(text => "your ip is $ip. This script adds iptables." );
};

app->start;
